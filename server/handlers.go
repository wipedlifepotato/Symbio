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
    "fmt"
    "sync"
    "math/big"
    "os"
    "mFrelance/models"
)
import "mFrelance/auth"

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

//
// POST /register
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
/////// WALLET
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

// Not Tested
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
		_ = json.Unmarshal(data, &pending) // если ошибка — пустой список
	}

	pending = append(pending, req)

	newData, _ := json.MarshalIndent(pending, "", "  ")
	return os.WriteFile(filePath, newData, 0644)
}

// TODO:
// Create a pool with transactions and send every N minutes the transaction with func (c *Client) PayToMany(outputs [][2]string) (string, error) {
// TODO: Tests
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

    var userBalanceStr, userAddress string
    err = db.Postgres.QueryRow(`SELECT balance::text, address FROM wallets WHERE user_id=$1 AND currency='BTC' LIMIT 1`,
        userID).Scan(&userBalanceStr, &userAddress)
    if err != nil {
        http.Error(w, "failed to get wallet: "+err.Error(), http.StatusInternalServerError)
        return
    }

    userBalance, _ := new(big.Float).SetString(userBalanceStr)
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
    newBalance := new(big.Float).Sub(userBalance, amount)
    newBalanceStr := fmt.Sprintf("%.8f", newBalance)
    _, err = db.Postgres.Exec(`UPDATE wallets SET balance=$1 WHERE user_id=$2 AND currency='BTC'`, newBalanceStr, userID)
    if err != nil {
        http.Error(w, "failed to update balance: "+err.Error(), http.StatusInternalServerError)
        return
    }

    isOur, err := db.IsOurAddr(db.Postgres, destAddress)
    if err != nil {
        http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
        return
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
	    _, err := db.Postgres.Exec(`
		UPDATE wallets 
		SET balance = balance + $1 
		WHERE address = $2 AND currency='BTC'`, remaining.Text('f', 8), destAddress)
	    if err != nil {
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
        "from":          userAddress,
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

/// PROfiles handlers
func ProfileHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        claims := GetUserFromContext(r)
        if claims == nil {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
	const MaxAvatarSize = 1 * 1024 * 1024 // 1 MB
        switch r.Method {
        case "GET":
            profile, err := models.GetProfile(db.Postgres, claims.UserID)
            if err != nil {
                http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
                return
            }
            json.NewEncoder(w).Encode(profile)
	case "POST":
	    var p models.Profile
	    if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	    }
	    p.UserID = claims.UserID

	    // Проверяем длину Base64 аватара
	    if len(p.Avatar) > MaxAvatarSize {
		http.Error(w, "avatar too large", http.StatusBadRequest)
		return
	    }

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

func ProfilesHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        profiles, err := models.GetAllProfiles(db.Postgres)
        if err != nil {
            http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
            return
        }
        json.NewEncoder(w).Encode(profiles)
    }
}
