package server
import (
    	"mFrelance/electrum"
    	"gitlab.com/moneropay/go-monero/walletrpc"
    	"time"
    	"context"
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"mFrelance/db"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"sync"
	"os"
	"encoding/json"
	"errors"
)

var txPool = struct {
	sync.Mutex
	outputs map[string]*big.Float
}{outputs: make(map[string]*big.Float)}


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
				if isTxProcessed(tx.Txid) {
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
					SaveTransaction(tx.Txid,walletID, amt, currency, true)
				}
			}

		case "XMR":
			// TODO: 
			log.Printf("XMR wallet sync not implemented yet (wallet %d)", walletID)
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
}

func isTxProcessed(txid string) bool {
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
/*
// TODO: use it instead save to file
func StartTxPoolFlusher(client *electrum.Client, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			txPool.Lock()
			if len(txPool.outputs) == 0 {
				txPool.Unlock()
				continue
			}


			var outs [][2]string
			for addr, amt := range txPool.outputs {
				outs = append(outs, [2]string{addr, amt.Text('f', 8)})
			}


			txPool.outputs = make(map[string]*big.Float)
			txPool.Unlock()


			txid, err := client.PayToMany(outs)
			if err != nil {
				log.Println("Ошибка PayToMany:", err)

				txPool.Lock()
				for _, o := range outs {
					val, _ := new(big.Float).SetString(o[1])
					if existing, ok := txPool.outputs[o[0]]; ok {
						txPool.outputs[o[0]] = new(big.Float).Add(existing, val)
					} else {
						txPool.outputs[o[0]] = val
					}
				}
				txPool.Unlock()
			} else {
				log.Println("PayToMany успешно, txid:", txid)
			}
		}
	}()
}*/
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
    Time    string     `json:"time"`
    Outputs [][2]string `json:"outputs"`
    FeeBTC  float64    `json:"fee_btc"`
    Error   string     `json:"error"`
}

var paymentMu sync.Mutex

var paymentFile = "pending_payments.json"

func clearPendingPayments() error {
	paymentMu.Lock()
	defer paymentMu.Unlock()
	return os.WriteFile(paymentFile, []byte("[]"), 0644)
}
func StartTxPoolFlusher(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			txPool.Lock()
			if len(txPool.outputs) == 0 {
				txPool.Unlock()
				continue
			}

			var outs [][2]string
			for addr, amt := range txPool.outputs {
				outs = append(outs, [2]string{addr, amt.Text('f', 8)})
			}

			txPool.outputs = make(map[string]*big.Float)
			txPool.Unlock()

			record := PaymentRecord{
				Time:    time.Now().Format(time.RFC3339),
				Outputs: outs,
				FeeBTC:  0,
				Error:   "pending", 
			}

			if err := savePendingPayment(record); err != nil {
				log.Println("Error to save pending payment:", err)
			} else {
				log.Printf("create pending payment: %+v\n", record)
			}
		}
	}()
}
