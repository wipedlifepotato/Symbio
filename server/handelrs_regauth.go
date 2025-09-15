package server
import (
	"context"
    "net/http"
    "strconv"
    "github.com/go-redis/redis/v8"
    "math/rand"
    "github.com/steambap/captcha"
    "time"
    "encoding/json"
    "mFrelance/db"
	"mFrelance/auth"
	"log"
)
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

