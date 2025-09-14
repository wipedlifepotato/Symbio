package server

import (
    //"bytes"
    "encoding/json"
    "github.com/go-redis/redis/v8"
    //"io"
    "github.com/steambap/captcha"
    "log"
    "mFrelance/db"
    "math/rand"
    "net/http"
    //"os/exec"
    "strconv"
    "context"
    "time"
    "mFrelance/electrum"
    "gitlab.com/moneropay/go-monero/walletrpc"
    "mFrelance/config"
   // "fmt"
    "sync"
    "math/big"
    "os"
    "mFrelance/models"
    "database/sql"
)
import "github.com/lib/pq"
import "mFrelance/auth"
//import "io"
type RegisterRequest struct {
    Username      string `json:"username"`
    Password      string `json:"password"`
    // GPGKey        string `json:"gpg_key"`
    CaptchaID     string `json:"captcha_id"`
    CaptchaAnswer string `json:"captcha_answer"`
}


type Response struct {
    Message string `json:"message"`
    Encrypted string `json:"encrypted,omitempty"`
}




func HelloHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    resp := Response{Message: "Hello, REST API!"}
    json.NewEncoder(w).Encode(resp)
}

var ctx = context.Background()

func GetCaptchaFromRedis(rdb *redis.Client, id string) (string, error) {
    val, err := rdb.Get(ctx, "captcha:"+id).Result()
    if err != nil {
        return "", err
    }
    return val, nil
}

func CaptchaHandler(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {
    data, _ := captcha.New(150, 50)
    id := strconv.Itoa(rand.Int())
    rdb.Set(ctx, "captcha:"+id, data.Text, 5*time.Minute)

    w.Header().Set("Content-Type", "image/png")
    w.Header().Set("X-Captcha-ID", id)
    data.WriteImage(w)
}

func VerifyHandler(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {
    id := r.URL.Query().Get("id")
    answer := r.URL.Query().Get("answer")

    stored, err := rdb.Get(ctx, "captcha:"+id).Result()
    //log.Print("[DEBUG, Handlers] Captcha: "+id+"="+stored)
    if err != nil {
        writeErrorJSON(w, "Captcha expired", http.StatusBadRequest)
        return
    }

    if answer == stored {
        rdb.Del(ctx, "captcha:"+id)
        w.Write([]byte(`{"ok":true}`))
        return
    }

    w.Write([]byte(`{"ok":false}`))
}

// RegisterHandler godoc
// @Summary Register new user
// @Description Creates a new user with login, password and captcha
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "User credentials"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /register [post]
func RegisterHandler(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {
    var req RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeErrorJSON(w, "invalid json", http.StatusBadRequest)
        return
    }

    storedCaptcha, err := rdb.Get(ctx, "captcha:"+req.CaptchaID).Result()
    if err != nil || storedCaptcha != req.CaptchaAnswer {
        writeErrorJSON(w, "invalid captcha", http.StatusBadRequest)
        return
    }
    rdb.Del(ctx, "captcha:"+req.CaptchaID)


    mnemonic := GenerateMnemonic()

    passwordHash := HashPassword(req.Password)

// func CreateUser(db *sqlx.DB, username, passwordHash, gpgKey string) error {
    err = db.CreateUser(db.Postgres, req.Username, passwordHash, mnemonic)
    if err != nil {
        writeErrorJSON(w, "failed to create user", http.StatusInternalServerError)
        return
    }

    resp := Response{
        Message:   "Account created successfully. Save your recovery phrase!",
        Encrypted: mnemonic,
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}
type RestoreRequest struct {
	Username      string `json:"username"`
	Mnemonic      string `json:"mnemonic"`
	NewPassword   string `json:"new_password"`
        CaptchaID     string `json:"captcha_id"`
        CaptchaAnswer string `json:"captcha_answer"`
}
// POST /restoreuser
func RestoreHandler(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {
  var req RestoreRequest
  if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeErrorJSON(w, "invalid json", http.StatusBadRequest)
        return
  }  
  storedCaptcha, err := rdb.Get(ctx, "captcha:"+req.CaptchaID).Result()
  if err != nil || storedCaptcha != req.CaptchaAnswer {
        writeErrorJSON(w, "invalid captcha", http.StatusBadRequest)
        return
  }
  rdb.Del(ctx, "captcha:"+req.CaptchaID)
  // func RestoreUser(db *sqlx.DB, wantusername, mnemonic string) (int64, string, error) {
  userID, username, err := db.RestoreUser(db.Postgres, req.Username, req.Mnemonic)
    if err != nil || userID == 0 || username == "" {
        writeErrorJSON(w, "failed to found user", http.StatusInternalServerError)
        return
  }
  passwordHash := HashPassword(req.NewPassword)
  // func ChangeUserPassword(db *sqlx.DB, username, passwordHash string) error {
  err = db.ChangeUserPassword(db.Postgres, req.Username, passwordHash)
  if err != nil {
  	writeErrorJSON(w, "Failed to change user password", http.StatusInternalServerError)
  	return
  }
  token, err := auth.GenerateJWT(userID, username)
  if err != nil {
		writeErrorJSON(w, "Failed to generate token", http.StatusInternalServerError)
		return
  }
  resp := Response{
        Message:   "Account restored successfully.",
        Encrypted: token,
  }
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(resp)  
}

///

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	CaptchaID string `json:"captcha_id"`
	CaptchaAnswer string `json:"captcha_answer"`
}

type AuthResponse struct {
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
}

// AuthHandler godoc
// @Summary Authenticate user
// @Description Logs in user and returns JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body AuthRequest true "Login credentials"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth [post]
func AuthHandler(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorJSON(w, "invalid json", http.StatusBadRequest)
		return
	}
	log.Print("[AuthHandler] Test captcha")
	storedCaptcha, err := rdb.Get(ctx, "captcha:"+req.CaptchaID).Result()
	if err != nil || storedCaptcha != req.CaptchaAnswer {
		writeErrorJSON(w, "invalid captcha", http.StatusBadRequest)
		return
	}
	rdb.Del(ctx, "captcha:"+req.CaptchaID)

	log.Print("[AuthHandler] Get User by Username")
	userID, passwordHash, err := db.GetUserByUsername(db.Postgres, req.Username)
	if err != nil {
		writeErrorJSON(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if userID == 0 {
		writeErrorJSON(w, "invalid username or password", http.StatusUnauthorized)
		return
	}
	log.Print("[AuthHandler] Check Password")
	if !VerifyPassword(req.Password, passwordHash) {
		writeErrorJSON(w, "invalid username or password", http.StatusUnauthorized)
		return
	}
	log.Print("[AuthHandler] Generate JWT")
	token, err := auth.GenerateJWT(userID, req.Username)
	if err != nil {
		writeErrorJSON(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	resp := AuthResponse{
		Message: "Authenticated successfully",
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func TestHandler(w http.ResponseWriter, r *http.Request) {
	claims := GetUserFromContext(r)
	if claims == nil {
		writeErrorJSON(w, "user not found in context", http.StatusUnauthorized)
		return
	}

	resp := map[string]any{
		"user_id": claims.UserID,
		"username": claims.Username,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// WalletHandler godoc
// @Summary Get wallet balances
// @Description Returns user’s balances in BTC and XMR
// @Tags wallet
// @Produce json
// @Success 200 {object} db.WalletBalance
// @Security BearerAuth
// @Router /api/wallet [get]
func WalletHandler(w http.ResponseWriter, r *http.Request, mClient *walletrpc.Client, eClient *electrum.Client) {
	claims := GetUserFromContext(r)
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
			resp, err := mClient.CreateAddress(context.Background(), &walletrpc.CreateAddressRequest{
				AccountIndex: 0,
				Label:        "user_" + strconv.Itoa(int(userID)),
			})
			if err != nil {
				http.Error(w, "Monero RPC error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			address = resp.Address
		case "BTC":
			addr, err := eClient.CreateAddress()
			if err != nil {
				http.Error(w, "Electrum error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			address = addr
		default:
			http.Error(w, "Unsupported currency", http.StatusBadRequest)
			return
		}

		_, err = db.Postgres.Exec(`INSERT INTO wallets(user_id,currency,address) VALUES($1,$2,$3)`, userID, currency, address)
		if err != nil {
			http.Error(w, "DB insert error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		wallet = &db.WalletBalance{
			Address: address,
			Balance: 0,
		}
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
	// TODO:
}


var txPoolBlocked struct {
	sync.RWMutex
	Blocked bool
}

func SetTxPoolBlocked(block bool) {
	txPoolBlocked.Lock()
	defer txPoolBlocked.Unlock()
	txPoolBlocked.Blocked = block
}

func IsTxPoolBlocked() bool {
	txPoolBlocked.RLock()
	defer txPoolBlocked.RUnlock()
	return txPoolBlocked.Blocked
}

type PendingRequest struct {
	UserID      int64   `json:"user_id"`
	To          string  `json:"to"`
	Amount      string  `json:"amount"`
	Commission  string  `json:"commission"`
	Remaining   string  `json:"remaining"`
	Timestamp   int64   `json:"timestamp"`
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
	if IsTxPoolBlocked() {
		http.Error(w, "withdrawals temporarily blocked", http.StatusForbidden)
		return
	}

	claims := GetUserFromContext(r)
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

	amount, ok := new(big.Float).SetString(amountStr)
	if !ok {
		http.Error(w, "invalid amount", http.StatusBadRequest)
		return
	}

	if !IsValidBTCAddress(destAddress) {
		http.Error(w, "invalid Bitcoin address format", http.StatusBadRequest)
		return
	}

	minBTC := big.NewFloat(0.0001)
	if amount.Cmp(minBTC) < 0 {
		http.Error(w, "amount below minimum", http.StatusBadRequest)
		return
	}

	userWallet, err := models.GetWalletByUserAndCurrency(userID, "BTC")
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

	req := PendingRequest{
		UserID:     userID,
		To:         destAddress,
		Amount:     amountStr,
		Commission: commission.Text('f', 8),
		Remaining:  remaining.Text('f', 8),
		Timestamp:  time.Now().Unix(),
	}
	if err := savePendingRequest(req); err != nil {
		log.Printf("Failed to save pending request: %v", err)
	}

	if err := userWallet.SubBalance(amount); err != nil {
		http.Error(w, "failed to update balance: "+err.Error(), http.StatusInternalServerError)
		return
	}

	isOur, err := models.IsOurWalletAddress(destAddress, "BTC")
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tx := models.Transaction{
		FromWalletID: sql.NullInt64{Int64: userWallet.ID, Valid: true},
		ToWalletID:   sql.NullInt64{Valid: false},
		ToAddress:    sql.NullString{String: destAddress, Valid: true},
		Amount:       remaining.Text('f', 8),
		Currency:     "BTC",
		Confirmed:    false,
	}

	if isOur {
		if destWallet, err := models.GetWalletByAddress(destAddress, "BTC"); err == nil {
			tx.ToWalletID = sql.NullInt64{Int64: destWallet.ID, Valid: true}
		}
	}

	if err := models.SaveTransaction(&tx); err != nil {
		log.Printf("Failed to save transaction: %v", err)
	}

	if !isOur {
		txPool.Lock()
		if existing, ok := txPool.outputs[config.AppConfig.BitcoinAddress]; ok {
			txPool.outputs[config.AppConfig.BitcoinAddress] = new(big.Float).Add(existing, commission)
		} else {
			txPool.outputs[config.AppConfig.BitcoinAddress] = new(big.Float).Set(commission)
		}
		if existing, ok := txPool.outputs[destAddress]; ok {
			txPool.outputs[destAddress] = new(big.Float).Add(existing, remaining)
		} else {
			txPool.outputs[destAddress] = new(big.Float).Set(remaining)
		}
		txPool.Unlock()

		log.Printf("Added to pool: to=%s amount=%s commission=%s", destAddress, remaining.Text('f', 8), commission.Text('f', 8))
	} else {
		if err := models.AddToWalletBalance(destAddress, "BTC", remaining); err != nil {
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

///// Admin Handlers

type AdminRequest struct {
	UserID int64 `json:"user_id"`
}


func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := GetUserFromContext(r)
		if claims == nil {
			http.Error(w, "user not found (RequireAdmin)", http.StatusUnauthorized)
			return
		}
		isAdmin, err := db.IsAdmin(db.Postgres, claims.UserID)
		if err != nil || !isAdmin {
			http.Error(w, "admin rights required", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

// MakeAdminHandler godoc
// @Summary Grant admin rights
// @Description Makes a user admin by userID
// @Tags admin
// @Accept json
// @Produce json
// @Param request body AdminRequest true "UserID payload"
// @Success 200 {string} string "user is now admin"
// @Failure 400 {string} string "invalid request body"
// @Failure 401 {string} string "unauthorized"
// @Failure 403 {string} string "admin rights required"
// @Failure 500 {string} string "internal server error"
// @Security BearerAuth
// @Router /api/admin/make [post]
func MakeAdminHandler(w http.ResponseWriter, r *http.Request) {
	var req AdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := db.MakeAdmin(db.Postgres, req.UserID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("user is now admin"))
}
// RemoveAdminHandler godoc
// @Summary Revoke admin rights
// @Description Removes admin status from a user
// @Tags admin
// @Accept json
// @Produce json
// @Param request body AdminRequest true "UserID payload"
// @Success 200 {string} string "user admin removed"
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Security BearerAuth
// @Router /api/admin/remove [post]
func RemoveAdminHandler(w http.ResponseWriter, r *http.Request) {
	var req AdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := db.RemoveAdmin(db.Postgres, req.UserID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("user admin removed"))
}

// IsAdminHandler godoc
// @Summary Check if user is admin
// @Description Returns true/false if current user has admin privileges
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Router /api/admin/check [get]
func IsAdminHandler(w http.ResponseWriter, r *http.Request) {
	var req AdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	isAdmin, err := db.IsAdmin(db.Postgres, req.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"user_id":  req.UserID,
		"is_admin": isAdmin,
	})
}

// BlockUserHandler godoc
// @Summary Block user
// @Description Blocks a user by userID
// @Tags admin
// @Accept json
// @Produce json
// @Param request body AdminRequest true "UserID payload"
// @Success 200 {string} string "user blocked"
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Security BearerAuth
// @Router /api/admin/block [post]
func BlockUserHandler(w http.ResponseWriter, r *http.Request) {
	var req AdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := db.BlockUser(db.Postgres, req.UserID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("user blocked"))
}
// AdminTransactionsHandler godoc
// @Summary Admin: View transactions
// @Description Allows admin to view transactions by wallet or all transactions with pagination
// @Tags admin
// @Accept json
// @Produce json
// @Param request body AdminTransactionsRequest true "Request payload"
// @Success 200 {array} object
// @Success 200 {array} object "id:int, from_wallet_id:int, to_wallet_id:int, to_address:string, task_id:int, amount:string, currency:string, confirmed:bool, created_at:string"
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Security BearerAuth
// @Router /api/admin/transactions [post]
func AdminTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	var req AdminTransactionsRequest
	//log.Println(r.Body)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	limit := req.Limit
	offset := req.Offset
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var txs []*models.Transaction
	var err error

	if req.WalletID > 0 {
		txs, err = models.GetTransactionsByWallet(int64(req.WalletID), limit, offset)
		if err != nil {
			http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		txs, err = models.GetTransactions(limit, offset)
		if err != nil {
			http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(txs)
}

type AdminTransactionsRequest struct {
	WalletID int `json:"wallet_id,omitempty"`
	Limit    int `json:"limit,omitempty"`
	Offset   int `json:"offset,omitempty"`
}

// UnblockUserHandler godoc
// @Summary Unblock user
// @Description Unblocks a user by userID
// @Tags admin
// @Accept json
// @Produce json
// @Param request body AdminRequest true "UserID payload"
// @Success 200 {string} string "user unblocked"
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Security BearerAuth
// @Router /api/admin/unblock [post]
func UnblockUserHandler(w http.ResponseWriter, r *http.Request) {
	var req AdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := db.UnblockUser(db.Postgres, req.UserID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("user unblocked"))
}

// ProfileHandler godoc
// @Summary Get or update profile
// @Description Get current user profile (GET) or update profile (POST)
// @Tags profile
// @Accept json
// @Produce json
// @Success 200 {object} models.Profile
// @Failure 400 {string} string "invalid payload"
// @Failure 401 {string} string "unauthorized"
// @Failure 500 {string} string "db error"
// @Security BearerAuth
// @Router /profile [get]
// @Router /profile [post]
func ProfileHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        claims := GetUserFromContext(r)
        if claims == nil {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
	maxAvatarSize := int(config.AppConfig.MaxAvatarSize) * 1024 * 1024 // MBs
        switch r.Method {
        case "GET":
            profile, err := models.GetProfile(db.Postgres, claims.UserID)
            if err != nil {
                http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
                return
            }
            SanitizeProfile(profile)
            json.NewEncoder(w).Encode(profile)
	case "POST":
		var p models.Profile
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		    log.Println("Decode error:", err)
		    http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
		    return
		}
	    p.UserID = claims.UserID

	    if len(p.Avatar) > maxAvatarSize {
		http.Error(w, "avatar too large", http.StatusBadRequest)
		return
	    }
	    SanitizeProfile(&p)
	    if err := models.UpsertProfile(db.Postgres, &p); err != nil {
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	    }
	    w.WriteHeader(http.StatusOK)
	    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

        default:
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        }
    }
}

// ProfilesHandler godoc
// @Summary List profiles
// @Description Returns paginated list of profiles
// @Tags profile
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} models.Profile
// @Failure 500 {string} string "db error"
// @Security BearerAuth
// @Router /profiles [get]
func ProfilesHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {

        limit := 50
        offset := 0

        if l := r.URL.Query().Get("limit"); l != "" {
            if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 500 {
                limit = v
            } else if err != nil {
            	//log.Println(err)
            }
        } else {
        	//log.Println("LIMIT IS ''")
        }
        if limit > int(config.AppConfig.MaxProfiles) {
   		     limit = int(config.AppConfig.MaxProfiles);
        }
        if o := r.URL.Query().Get("offset"); o != "" {
            if v, err := strconv.Atoi(o); err == nil && v >= 0 {
                offset = v
            }
        }
	//log.Println("Get profiles: ")
	//log.Println(limit)
	//log.Println(offset)
        profiles, err := models.GetProfilesWithLimitOffset(db.Postgres, limit, offset)
        if err != nil {
            http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
            return
        }

        json.NewEncoder(w).Encode(profiles)
    }
}


// AdminWalletsHandler godoc
// @Summary Get user wallets
// @Description Returns all wallets for a given user
// @Tags admin
// @Accept json
// @Produce json
// @Param user_id query int true "User ID"
// @Success 200 {array} models.Wallet
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Security BearerAuth
// @Router /api/admin/wallets [get]
func AdminWalletsHandler(w http.ResponseWriter, r *http.Request) {
    userIDStr := r.URL.Query().Get("user_id")
    if userIDStr == "" {
        http.Error(w, "user_id required", http.StatusBadRequest)
        return
    }
    userID, err := strconv.ParseInt(userIDStr, 10, 64)
    if err != nil {
        http.Error(w, "invalid user_id", http.StatusBadRequest)
        return
    }

    wallets, err := models.GetWalletsByUser(userID)
    if err != nil {
        http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(wallets)
}
// TicketCreateRequest represents request to create ticket
type TicketCreateRequest struct {
	Message string `json:"message"`
	Subject string `json:"subject"`
}
// TicketCreateAnswer represents response after creating ticket
type TicketCreateAnswer struct {
	TicketID int64 `json:"ticket_id"`
}
// CreateTicket godoc
// @Summary Create new ticket
// @Description Create new ticket
// @Tags ticket
// @Accept  json
// @Produce  json
// @Param request body TicketCreateRequest true "Ticket info"
// @Success 200 {object} TicketCreateAnswer
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/ticket/createTicket [post]
// @Security bearerAuth
func CreateTicket(w http.ResponseWriter, r *http.Request) {
	claims := GetUserFromContext(r)
	if claims == nil {
		writeErrorJSON(w, "user not found in context", http.StatusUnauthorized)
		return
	}

	var t TicketCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		log.Println("Decode error:", err)
		http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	err := ValidateTicketField(t.Subject,t.Message)
	if err != nil {
		writeErrorJSON(w, "invalid parameters for ticket: "+err.Error(), http.StatusBadRequest)
		return
	}

	id, err := models.CreateTicket(SanitizeString(t.Subject), claims.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = models.AddTicketMessage(id, claims.UserID, SanitizeString(t.Message))
	if err != nil {
		writeErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TicketCreateAnswer{TicketID: id})
}

type WriteTicketRequest struct {
    TicketID int64  `json:"ticket_id"`
    Message  string `json:"message"`
}

// WriteToTicketHandler godoc
// @Summary Write to ticket
// @Description Add message to ticket
// @Tags ticket
// @Accept  json
// @Produce  json
// @Param request body WriteTicketRequest true "Message info"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/ticket/write [post]
// @Security bearerAuth
func WriteToTicketHandler(w http.ResponseWriter, r *http.Request) {
    claims := GetUserFromContext(r)
    if claims == nil {
        writeErrorJSON(w, "user not found in context", http.StatusUnauthorized)
        return
    }
    var req WriteTicketRequest

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeErrorJSON(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
        return
    }

    var ticket struct {
        UserID          *int64  `db:"user_id"`
        AdminID         *int64  `db:"admin_id"`
        AdditionalUsers pq.Int64Array `db:"additional_users_have_access"`
    }

    err := db.Postgres.Get(&ticket, `
        SELECT user_id, admin_id, additional_users_have_access
        FROM tickets
        WHERE id=$1
    `, req.TicketID)
    if err != nil {
        writeErrorJSON(w, "ticket not found: "+err.Error(), http.StatusNotFound)
        return
    }

    hasAccess := false
    if ticket.UserID != nil && *ticket.UserID == claims.UserID {
        hasAccess = true
    }
    if ticket.AdminID != nil && *ticket.AdminID == claims.UserID {
        hasAccess = true
    }
    for _, id := range ticket.AdditionalUsers {
        if id == claims.UserID {
            hasAccess = true
            break
        }
    }

    if !hasAccess {
        writeErrorJSON(w, "you do not have access to this ticket", http.StatusForbidden)
        return
    }

    if err := ValidateMessage(req.Message); err != nil {
        writeErrorJSON(w, "invalid message: "+err.Error(), http.StatusBadRequest)
        return
    }

    err = models.AddTicketMessage(req.TicketID, claims.UserID, req.Message)
    if err != nil {
        writeErrorJSON(w, "failed to add message: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

type TicketIDRequest struct {
    TicketID int64 `json:"ticket_id"`
}

// ExitFromTicketHandler godoc
// @Summary Exit from a ticket
// @Description Removes the user from the ticket's participants
// @Tags ticket
// @Accept json
// @Produce json
// @Param request body TicketIDRequest true "Ticket ID"
// @Success 200 {object} map[string]string "status: ok"
// @Failure 400 {object} map[string]string "Invalid payload or ticket_id"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/ticket/exit [post]
// @Security bearerAuth
func ExitFromTicketHandler(w http.ResponseWriter, r *http.Request) {
    claims := GetUserFromContext(r)
    if claims == nil {
        writeErrorJSON(w, "user not found in context", http.StatusUnauthorized)
        return
    }

    var req struct {
        TicketID int64 `json:"ticket_id"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeErrorJSON(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
        return
    }

    if req.TicketID == 0 {
        writeErrorJSON(w, "ticket_id is required", http.StatusBadRequest)
        return
    }

    err := models.ExitFromTicket(req.TicketID, claims.UserID)
    if err != nil {
        writeErrorJSON(w, "failed to exit from ticket: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
// GetTicketMessagesHandler godoc
// @Summary Get messages for a ticket
// @Description Returns all messages for a given ticket if the user has access
// @Tags ticket
// @Produce json
// @Param ticket_id query int true "Ticket ID"
// @Success 200 {array} models.TicketMessage "List of messages"
// @Failure 400 {object} map[string]string "Invalid ticket_id"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 403 {object} map[string]string "User does not have access"
// @Router /api/ticket/messages [get]
// @Security bearerAuth
func GetTicketMessagesHandler(w http.ResponseWriter, r *http.Request) {
    claims := GetUserFromContext(r)
    if claims == nil {
        writeErrorJSON(w, "user not found in context", http.StatusUnauthorized)
        return
    }

    ticketIDStr := r.URL.Query().Get("ticket_id")
    if ticketIDStr == "" {
        writeErrorJSON(w, "ticket_id is required", http.StatusBadRequest)
        return
    }

    ticketID, err := strconv.ParseInt(ticketIDStr, 10, 64)
    if err != nil {
        writeErrorJSON(w, "invalid ticket_id", http.StatusBadRequest)
        return
    }

    messages, err := models.GetMessagesForTicket(ticketID, claims.UserID)
    if err != nil {
        writeErrorJSON(w, err.Error(), http.StatusForbidden)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(messages)
}




// GetMyTicketsHandler godoc
// @Summary Get own tickets
// @Description Get all tickets of user
// @Tags ticket
// @Produce json
// @Success 200 {array} models.TicketDoc
// @Failure 401 {object} map[string]string
// @Router /api/ticket/my [get]
// @Security bearerAuth
func GetMyTicketsHandler(w http.ResponseWriter, r *http.Request) {
    claims := GetUserFromContext(r)
    if claims == nil {
        writeErrorJSON(w, "user not found in context", http.StatusUnauthorized)
        return
    }

    tickets, err := models.GetTicketsForUser(claims.UserID)
    if err != nil {
        writeErrorJSON(w, "failed to get tickets: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(tickets)
}

func CloseTicketHandler(w http.ResponseWriter, r *http.Request) {
    claims := GetUserFromContext(r)
    if claims == nil {
        writeErrorJSON(w, "user not found in context", http.StatusUnauthorized)
        return
    }

    var req struct {
        TicketID int64 `json:"ticket_id"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeErrorJSON(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
        return
    }

    if req.TicketID == 0 {
        writeErrorJSON(w, "ticket_id is required", http.StatusBadRequest)
        return
    }

	var ticket struct {
		UserID          *int64        `db:"user_id"`
		AdminID         *int64        `db:"admin_id"`
		AdditionalUsers pq.Int64Array `db:"additional_users_have_access"`
	}

    err := db.Postgres.Get(&ticket, `
        SELECT user_id, admin_id, additional_users_have_access
        FROM tickets
        WHERE id=$1
    `, req.TicketID)
    if err != nil {
        writeErrorJSON(w, "ticket not found: "+err.Error(), http.StatusNotFound)
        return
    }

    allowed := false
    if ticket.UserID != nil && *ticket.UserID == claims.UserID {
        allowed = true
    }
    if ticket.AdminID != nil && *ticket.AdminID == claims.UserID {
        allowed = true
    }
    for _, id := range ticket.AdditionalUsers {
        if id == claims.UserID {
            allowed = true
            break
        }
    }

    if !allowed {
        writeErrorJSON(w, "access denied: you are not part of this ticket", http.StatusForbidden)
        return
    }

    if err := models.CloseTicket(req.TicketID); err != nil {
        writeErrorJSON(w, "failed to close ticket: "+err.Error(), http.StatusInternalServerError)
        return
    }

    if err := models.ExitFromTicket(req.TicketID, claims.UserID); err != nil {
        writeErrorJSON(w, "ticket closed, but failed to exit from ticket: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "closed"})
}

// AdminGetRandomTicketHandler godoc
// @Summary Get random opened ticket (admin)
// @Description Set ticket to admin (random)
// @Tags admin
// @Produce json
// @Success 200 {object} models.TicketDoc
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/admin/getRandomTicket [get]
// @Security bearerAuth
func AdminGetRandomTicketHandler(w http.ResponseWriter, r *http.Request) {
	claims := GetUserFromContext(r)
	if claims == nil {
		writeErrorJSON(w, "user not found in context", http.StatusUnauthorized)
		return
	}

    ticket,err := models.GetRandomOpenTicket()
	if err != nil {
		writeErrorJSON(w,err.Error(), http.StatusBadRequest)
		return
	}
	err = models.AssignTicketAdmin(ticket.ID, claims.UserID)
	if err != nil {
		writeErrorJSON(w, err.Error(), http.StatusBadRequest)
	}
	err = models.PendingTicket(ticket.ID)
	if err != nil {
		writeErrorJSON(w, err.Error(), http.StatusBadRequest)
	}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(ticket)
}

type AdminUpdateBalanceRequest struct {
    UserID int64  `json:"user_id"`
    Balance  string `json:"balance"` 
}

// AdminUpdateBalanceHandler godoc
// @Summary Update wallet balance
// @Description Allows admin to set a new balance for a wallet
// @Tags admin
// @Accept json
// @Produce json
// @Param request body AdminUpdateBalanceRequest true "Wallet balance payload"
// @Success 200 {string} string "balance updated"
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Security BearerAuth
// @Router /api/admin/update_balance [post]
func AdminUpdateBalanceHandler(w http.ResponseWriter, r *http.Request) {
    var req AdminUpdateBalanceRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }

    newBalance, ok := new(big.Float).SetString(req.Balance)
    if !ok {
        http.Error(w, "invalid balance format", http.StatusBadRequest)
        return
    }
    //log.Println("Req wallet ID")
    //log.Println(req.WalletID)
    _, err := db.Postgres.Exec(
        `UPDATE wallets SET balance=$1 WHERE user_id=$2`,
        newBalance.Text('f', 12),
        req.UserID,
    )
    if err != nil {
        http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("balance updated"))
}
