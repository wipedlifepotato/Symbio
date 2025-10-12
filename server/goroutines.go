package server

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"gitlab.com/moneropay/go-monero/walletrpc"
	"log"
	"mFrelance/config"
	"mFrelance/db"
	"mFrelance/electrum"
	"math/big"
	"os"
	"sync"
	"time"
)

type TxPoolItem struct {
    Amount   *big.Float
    Currency string
}

var txPool = struct {
    sync.Mutex
    outputs map[string][]TxPoolItem 
}{outputs: make(map[string][]TxPoolItem)}

var lastTx string

func ProcessTxElectrum(client *electrum.Client, address string, txHash string) (*big.Float, error) {

	confirmations, err := client.Get_tx_status(txHash)
	if err != nil {
		return nil, err
	}

	if confirmations < 1 {
		return nil, errors.New("транзакция ещё не подтверждена")
	}

	txHexRaw, err := client.GetTransaction(txHash)
	if err != nil {
		return nil, err
	}

	rawTx, err := hex.DecodeString(txHexRaw)
	if err != nil {
		return nil, err
	}

	msgTx := wire.NewMsgTx(wire.TxVersion)
	if err := msgTx.Deserialize(bytes.NewReader(rawTx)); err != nil {
		return nil, err
	}

	incoming := big.NewFloat(0)

	for _, txOut := range msgTx.TxOut {
		_, addrs, _, err := txscript.ExtractPkScriptAddrs(txOut.PkScript, &chaincfg.TestNet3Params)
		if err != nil || len(addrs) == 0 {
			continue
		}

		for _, addr := range addrs {
			//log.Print(addr.EncodeAddress())
			//log.Print(txOut.Value)
			if addr.EncodeAddress() == address {
				val := new(big.Float).Quo(big.NewFloat(float64(txOut.Value)), big.NewFloat(1e8))
				incoming.Add(incoming, val)
			}
		}
	}

	return incoming, nil
}

func StartWalletSync(ctx context.Context, eClient *electrum.Client, mClient *walletrpc.Client, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("Wallet sync stopped")
				return
			case <-ticker.C:
				syncAllWallets(eClient, mClient)
			}
		}
	}()
}

type MoneroWalletSubAddr struct {
    WalletID   int
    Currency   string
    Address    string
    BalanceStr string
}

func getMoneroWalletSubAddr(address, currency string) (*MoneroWalletSubAddr, error) {
    var w MoneroWalletSubAddr

    err := db.Postgres.QueryRow(
        `SELECT id, currency, address, balance FROM wallets WHERE address=$1 AND currency=$2`,
        address, currency,
    ).Scan(&w.WalletID, &w.Currency, &w.Address, &w.BalanceStr)

    if err != nil {
     //   log.Printf("Wallet not found for address %s: %v", address, err)
        return nil, err
    }

    return &w, nil
}

func syncMoneroWallets(mClient *walletrpc.Client) {
	ctx := context.Background()

	resp, err := mClient.GetTransfers(ctx, &walletrpc.GetTransfersRequest{
		AccountIndex: 0,
		In:           true, 
		Pending:      false,
		Failed:       false,
		Pool:         false,
	})
	if err != nil {
		log.Printf("Failed GetTransfers: %v", err)
		return
	}

	for _, tx := range resp.In {

		if tx.Confirmations < 10 {
			continue
		}

		if IsTxProcessed(tx.Txid) {
			continue
		}

		addrResp, err := mClient.GetAddress(ctx, &walletrpc.GetAddressRequest{
			AccountIndex: tx.SubaddrIndex.Major,
		})
		if err != nil || int(tx.SubaddrIndex.Minor) >= len(addrResp.Addresses) {
			log.Printf("Failed to get address for TX %s: %v", tx.Txid, err)
			continue
		}
		walletAddress := addrResp.Addresses[tx.SubaddrIndex.Minor].Address

		w, err := getMoneroWalletSubAddr(walletAddress, "XMR")
		if err != nil {
			SaveTransaction(tx.Txid, 0, big.NewFloat(0), "XMR", true)
			continue
		}
		println("Found an new transaction for {}:{}", w.Address, tx.Amount)

		currentBalance, _ := new(big.Float).SetString(w.BalanceStr)

		amt := new(big.Float).Quo(new(big.Float).SetFloat64(float64(tx.Amount)), big.NewFloat(1e12))
		newBalance := new(big.Float).Add(currentBalance, amt)

		if err := SaveTransaction(tx.Txid, w.WalletID, amt, "XMR", true); err != nil {
			log.Printf("Failed to save transaction %s: %v", tx.Txid, err)
			continue
		}

		newBalanceStr := fmt.Sprintf("%.12f", newBalance)
		_, err = db.Postgres.Exec(`UPDATE wallets SET balance=$1 WHERE id=$2`, newBalanceStr, w.WalletID)
		if err != nil {
			log.Printf("Failed to update balance for wallet %d: %v", w.WalletID, err)
			continue
		}

		log.Printf("Updated XMR wallet %d (%s): +%s XMR", w.WalletID, w.Address, amt.Text('f', 12))
	}
}
func syncAllWallets(eClient *electrum.Client, mClient *walletrpc.Client) {
	//log.Println("SyncAllWallets")
	rows, err := db.Postgres.Query(`SELECT id, currency, address, balance FROM wallets`)
	if err != nil {
		log.Println("Failed to fetch wallets:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var walletID int
		var currency, address string
		var balanceStr string

		if err := rows.Scan(&walletID, &currency, &address, &balanceStr); err != nil {
			log.Println("Failed to scan wallet:", err)
			continue
		}

		currentBalance, _ := new(big.Float).SetString(balanceStr)
		newBalance := new(big.Float).Set(currentBalance)

		switch currency {
		case "BTC":
			txs, err := eClient.ListTransactions(address)
			if err != nil {
				log.Println(address)
				log.Printf("Electrum ListTransactions error for wallet %d: %v", walletID, err)
				continue
			}

			for _, tx := range txs {
				if IsTxProcessed(tx.Txid) {
					continue
				}

				amt, err := ProcessTxElectrum(eClient, address, tx.Txid)
				if err != nil {
					log.Printf("Failed ProcessTxElectrum for %s: %v", tx.Txid, err)
					continue
				}

				if amt.Cmp(big.NewFloat(0)) > 0 {
					log.Println("New balance")
					log.Print(amt)
					log.Print("WalletID: ")
					log.Print(walletID)
					newBalance.Add(newBalance, amt)
					//func SaveTransaction(txid string, walletID int, amount decimal.Decimal, currency string, confirmed bool) error {
					SaveTransaction(tx.Txid, walletID, amt, currency, true)
				}
			}

		case "XMR":
			//log.Printf("XMR wallet sync will be in an another block (wallet %d)", walletID)
		default:
			log.Printf("Unknown currency %s for wallet %d", currency, walletID)
		}

		newBalanceStr := fmt.Sprintf("%.12f", newBalance)
		_, err = db.Postgres.Exec(`UPDATE wallets SET balance=$1 WHERE id=$2`, newBalanceStr, walletID)
		if err != nil {
			log.Println("Failed to update wallet balance:", err)
		} else {
			log.Printf("Wallet %d (%s) balance updated: %s", walletID, currency, newBalanceStr)
		}
	}
	syncMoneroWallets(mClient);
}

func IsTxProcessed(txid string) bool {
	var exists bool
	err := db.Postgres.QueryRow(`
        SELECT EXISTS(SELECT 1 FROM wallet_transactions WHERE txid = $1)
    `, txid).Scan(&exists)
	if err != nil {
		log.Printf("isTxProcessed error: %v", err)
		return false
	}
	return exists
}

func SaveTransaction(txid string, walletID int, amount *big.Float, currency string, confirmed bool) error {
	amountStr := amount.Text('f', 12)
	_, err := db.Postgres.Exec(`
        INSERT INTO wallet_transactions (txid, wallet_id, amount, currency, confirmed, created_at)
        VALUES ($1, $2, $3, $4, $5, NOW())
        ON CONFLICT (txid) DO NOTHING
    `, txid, walletID, amountStr, currency, confirmed)
	return err
}

func savePendingPayment(record PaymentRecord) error {
	paymentMu.Lock()
	defer paymentMu.Unlock()

	var records []PaymentRecord

	data, err := os.ReadFile(paymentFile)
	if err == nil {
		_ = json.Unmarshal(data, &records)
	}

	records = append(records, record)

	fileData, _ := json.MarshalIndent(records, "", "  ")
	return os.WriteFile(paymentFile, fileData, 0644)
}

type PaymentRecord struct {
	Time    string      `json:"time"`
	Outputs [][2]string `json:"outputs"`
	FeeBTC  float64     `json:"fee_btc"`
	Error   string      `json:"error"`
}

var paymentMu sync.Mutex

var paymentFile = "pending_payments.json"

func clearPendingPayments() error {
	paymentMu.Lock()
	defer paymentMu.Unlock()
	return os.WriteFile(paymentFile, []byte("[]"), 0644)
}
func StartTxBlockTransactions(ctx context.Context, client *electrum.Client, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if lastTx == "" {
					continue
				}

				_, err := ProcessTxElectrum(client, config.AppConfig.BitcoinAddress, lastTx)
				if err != nil {
					log.Println("Tx check failed, blocking:", err)
					SetTxPoolBlocked(true)
				} else {
					log.Println("Tx check OK, unblocking. Resetting lastTx")
					SetTxPoolBlocked(false)
					lastTx = ""
				}
			}
		}
	}()
}

func flushMoneroTxPool(mClient *walletrpc.Client, outs []struct {
    Address string
    Amount  *big.Float
}) {
    ctx := context.Background()

    var dests []walletrpc.Destination
    for _, o := range outs {
        amtPi := new(big.Int)
        o.Amount.Mul(o.Amount, big.NewFloat(1e12)) 
        o.Amount.Int(amtPi) 
        //amtFloat, _ := o.Amount.Float64()
        dests = append(dests, walletrpc.Destination{
            Address: o.Address,
            Amount:  amtPi.Uint64(), // piconero
        })
    }
    if len(dests) == 0 {
        return
    }

    resp, err := mClient.Transfer(ctx, &walletrpc.TransferRequest{
        Destinations: dests, 
        AccountIndex: 0,
        RingSize:     16,
    })
    if err != nil {
        log.Printf("Monero PayToMany failed: %v", err)
        return
    }

    log.Printf("Monero transaction successfully sent. TXID: %s, Fee: %.12f XMR",
        resp.TxHash, float64(resp.Fee)/1e12)
}

func StartTxPoolFlusher(client *electrum.Client, mClient *walletrpc.Client, interval time.Duration, maxBatchSize int) {
    ticker := time.NewTicker(interval)

    flush := func() {
        txPool.Lock()
        if len(txPool.outputs) == 0 {
            txPool.Unlock()
            return
        }

        payments := make(map[string][][2]string) // currency -> list of [address, amount]
        for addr, items := range txPool.outputs {
            for _, item := range items {
                amtStr := item.Amount.Text('f', 8)
                payments[item.Currency] = append(payments[item.Currency], [2]string{addr, amtStr})
            }
        }

        txPool.outputs = make(map[string][]TxPoolItem)
        txPool.Unlock()

        for currency, outs := range payments {
            var txid string
            var err error
            switch currency {
            case "BTC":
                txid, err = client.PayToMany(outs)
            case "XMR":
                var moneroOuts []struct {
                    Address string
                    Amount  *big.Float
                }
                for _, o := range outs {
                    amt, _ := new(big.Float).SetString(o[1])
                    moneroOuts = append(moneroOuts, struct {
                        Address string
                        Amount  *big.Float
                    }{Address: o[0], Amount: amt})
                }

                flushMoneroTxPool(mClient, moneroOuts)
            default:
                log.Println("Unknown currency in txPool:", currency)
                continue
            }

            if err != nil {
                log.Println("Ошибка PayToMany для", currency, ":", err)

                txPool.Lock()
                for _, o := range outs {
                    val, _ := new(big.Float).SetString(o[1])
                    txPool.outputs[o[0]] = append(txPool.outputs[o[0]], TxPoolItem{Amount: val, Currency: currency})
                }
                txPool.Unlock()

                f, _ := os.OpenFile("FailedPayments.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
                defer f.Close()
                fmt.Fprintf(f, "%s - PayToMany %s failed: %v\nOutputs: %+v\n", time.Now().Format(time.RFC3339), currency, err, outs)
            } else {
                log.Println("PayToMany успешно для", currency, "txid:", txid)
                f, _ := os.OpenFile("SuccessfulPayments.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
                defer f.Close()
                fmt.Fprintf(f, "%s - %s txid: %s\nOutputs: %+v\n", time.Now().Format(time.RFC3339), currency, txid, outs)
            }
        }
    }

    go func() {
        for {
            select {
            case <-ticker.C:
                flush()
            default:
                txPool.Lock()
                tooBig := 0
                for _, items := range txPool.outputs {
                    tooBig += len(items)
                }
                txPool.Unlock()

                if tooBig >= maxBatchSize {
                    log.Println("TxPool достиг лимита, выполняю срочный сброс")
                    flush()
                } else {
                    time.Sleep(100 * time.Millisecond)
                }
            }
        }
    }()
}
