package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"time"
	"regexp"
	"gitlab.com/moneropay/go-monero/walletrpc"

	"mFrelance/config"
	"mFrelance/db"
	"mFrelance/electrum"
	"mFrelance/models"
	"mFrelance/server"
)

// WalletHandler godoc
// @Summary Get wallet balances
// @Description Returns user’s balances in BTC and XMR
// @Tags wallet
// @Produce json
// @Success 200 {object} db.WalletBalance
// @Security BearerAuth
// @Router /api/wallet [get]
func WalletHandler(w http.ResponseWriter, r *http.Request, mClient *walletrpc.Client, eClient *electrum.Client) {
	claims := server.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "user not found in context", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID
	currency := r.URL.Query().Get("currency")

	wallet, err := db.GetWalletBalance(db.Postgres, userID, currency)
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if wallet == nil {
		var address string
		switch currency {
		case "XMR":
			resp, err := mClient.CreateAddress(context.Background(), &walletrpc.CreateAddressRequest{AccountIndex: 0, Label: "user_" + strconv.Itoa(int(userID))})
			if err != nil {
				http.Error(w, "Monero RPC error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			address = resp.Address
		case "BTC":
			addr, err := eClient.CreateAddress(db.Postgres, "BTC")
			if err != nil {
				http.Error(w, "Electrum error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			address = addr
			txs, err := eClient.ListTransactions(address)
			if err != nil {
				http.Error(w, "Electrum ListTransactions error for new wallet", http.StatusInternalServerError)
				return
			}
			for _, tx := range txs {
				if server.IsTxProcessed(tx.Txid) {
					continue
				}
				//amt, err := server.ProcessTxElectrum(eClient, address, tx.Txid)
				//if err != nil {
				//		http.Error(w,"Failer to process tx",  http.StatusInternalServerError)
				//		continue
				//}
				err = server.SaveTransaction(tx.Txid, -1, big.NewFloat(0), currency, true)
				if err != nil {
					http.Error(w, "Failer to save transaction", http.StatusInternalServerError)
					return
				}
			}

		default:
			http.Error(w, "Unsupported currency", http.StatusBadRequest)
			return
		}
		_, err = db.Postgres.Exec(`INSERT INTO wallets(user_id,currency,address) VALUES($1,$2,$3)`, userID, currency, address)
		if err != nil {
			http.Error(w, "DB insert error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		wallet = &db.WalletBalance{Address: address, Balance: 0}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wallet)
}

// SendMoneroHandler godoc
// @Summary Send Monero
// @Description Sends Monero transaction (not implemented)
// @Tags wallet
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /api/wallet/moneroSend [post]
func SendMoneroHandler(w http.ResponseWriter, r *http.Request, mClient *walletrpc.Client) {
	// TODO
}

type PendingRequest struct {
	UserID     int64  `json:"user_id"`
	To         string `json:"to"`
	Amount     string `json:"amount"`
	Commission string `json:"commission"`
	Remaining  string `json:"remaining"`
	Timestamp  int64  `json:"timestamp"`
}

func savePendingRequest(req PendingRequest) error {
	const filePath = "pendingRequests.json"
	var pending []PendingRequest
	data, err := os.ReadFile(filePath)
	if err == nil {
		_ = json.Unmarshal(data, &pending)
	}
	pending = append(pending, req)
	newData, _ := json.MarshalIndent(pending, "", "  ")
	return os.WriteFile(filePath, newData, 0644)
}


var amountRe = regexp.MustCompile(`^\d+(\.\d{1,8})?$`) // 0.00100001 // 0.001000011 not will accept so no will lost money in prespective
// SendElectrumHandler godoc
// @Summary Send Bitcoin
// @Description Sends Bitcoin transaction using Electrum
// @Tags wallet
// @Accept json
// @Produce json
// @Param to query string true "Destination address"
// @Param amount query string true "Amount"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /api/wallet/bitcoinSend [post]
func SendElectrumHandler(w http.ResponseWriter, r *http.Request, client *electrum.Client) {
	if server.IsTxPoolBlocked() {
		http.Error(w, "withdrawals temporarily blocked", http.StatusForbidden)
		return
	}
	claims := server.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	blocked, err := db.IsUserBlocked(db.Postgres, userID)
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if blocked {
		http.Error(w, "user is blocked", http.StatusForbidden)
		return
	}

	destAddress := r.URL.Query().Get("to")
	amountStr := r.URL.Query().Get("amount")
	if destAddress == "" || amountStr == "" {
		http.Error(w, "destination and amount required", http.StatusBadRequest)
		return
	}
	if !amountRe.MatchString(amountStr) {
	    http.Error(w, "amount must have at most 8 decimal places", http.StatusBadRequest)
	    return
	}
	amount, ok := new(big.Float).SetString(amountStr)
	if !ok {
		http.Error(w, "invalid amount", http.StatusBadRequest)
		return
	}
	if !server.IsValidBTCAddress(destAddress) {
		http.Error(w, "invalid Bitcoin address format", http.StatusBadRequest)
		return
	}
	minBTC := big.NewFloat(0.001)
	if amount.Cmp(minBTC) < 0 {
		http.Error(w, "amount below minimum", http.StatusBadRequest)
		return
	}

	userWallet, err := models.GetWalletByUserAndCurrency(db.Postgres, userID, "BTC")
	if err != nil {
		http.Error(w, "failed to get wallet: "+err.Error(), http.StatusInternalServerError)
		return
	}
	userBalance := userWallet.BigBalance()
	if userBalance.Cmp(amount) < 0 {
		http.Error(w, "insufficient balance", http.StatusBadRequest)
		return
	}

	commissionPerc := big.NewFloat(config.AppConfig.BitcoinCommission)
	commission := new(big.Float).Quo(new(big.Float).Mul(amount, commissionPerc), big.NewFloat(100))
	remaining := new(big.Float).Sub(amount, commission)
	if remaining.Cmp(big.NewFloat(0)) <= 0 {
		http.Error(w, "amount too small for commission", http.StatusBadRequest)
		return
	}

	req := PendingRequest{UserID: userID, To: destAddress, Amount: amountStr, Commission: commission.Text('f', 8), Remaining: remaining.Text('f', 8), Timestamp: time.Now().Unix()}
	if err := savePendingRequest(req); err != nil {
		log.Printf("Failed to save pending request: %v", err)
	}

	if err := userWallet.SubBalance(db.Postgres, amount); err != nil {
		http.Error(w, "failed to update balance: "+err.Error(), http.StatusInternalServerError)
		return
	}

	isOur, err := db.IsOurWalletAddress(db.Postgres, destAddress, "BTC")
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tx := models.Transaction{FromWalletID: sql.NullInt64{Int64: userWallet.ID, Valid: true}, ToWalletID: sql.NullInt64{Valid: false}, ToAddress: sql.NullString{String: destAddress, Valid: true}, Amount: remaining.Text('f', 8), Currency: "BTC", Confirmed: false}
	if isOur {
		if destWallet, err := models.GetWalletByAddress(db.Postgres, destAddress, "BTC"); err == nil {
			tx.ToWalletID = sql.NullInt64{Int64: destWallet.ID, Valid: true}
		}
	}
	if err := db.SaveTransaction(db.Postgres, &tx); err != nil {
		log.Printf("Failed to save transaction: %v", err)
	}

	if !isOur {
		server.AddToTxPool(config.AppConfig.BitcoinAddress, commission)
		server.AddToTxPool(destAddress, remaining)
		log.Printf("Added to pool: to=%s amount=%s commission=%s", destAddress, remaining.Text('f', 8), commission.Text('f', 8))
	} else {
		if err := models.AddToWalletBalance(db.Postgres, destAddress, "BTC", remaining); err != nil {
			http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":        "ok",
		"queued_amount": amountStr,
		"commission":    commission.Text('f', 8),
		"remaining":     remaining.Text('f', 8),
		"from":          userWallet.Address,
		"to":            destAddress,
	})
}
