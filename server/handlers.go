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
   //"fmt"
   // "github.com/ProtonMail/go-crypto/openpgp"
    //"strings"
    "golang.org/x/crypto/bcrypt"
   // "crypto/rand"
   bip39 "github.com/tyler-smith/go-bip39"


)
func GenerateMnemonic() string {

    entropy, err := bip39.NewEntropy(128)
    if err != nil {
        log.Fatal(err)
    }


    mnemonic, err := bip39.NewMnemonic(entropy)
    if err != nil {
        log.Fatal(err)
    }

    return mnemonic
}


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


func hashPassword(password string) string {
    hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        log.Fatal("Failed to hash password:", err)
    }
    return string(hashed)
}


func verifyPassword(password, hashed string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
    return err == nil
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
        http.Error(w, "Captcha expired", http.StatusBadRequest)
        return
    }

    if answer == stored {
        rdb.Del(ctx, "captcha:"+id) // удалить после проверки
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
        http.Error(w, "invalid json", http.StatusBadRequest)
        return
    }

    storedCaptcha, err := rdb.Get(ctx, "captcha:"+req.CaptchaID).Result()
    if err != nil || storedCaptcha != req.CaptchaAnswer {
        http.Error(w, "invalid captcha", http.StatusBadRequest)
        return
    }
    rdb.Del(ctx, "captcha:"+req.CaptchaID)


    mnemonic := GenerateMnemonic()

    passwordHash := hashPassword(req.Password)

// func CreateUser(db *sqlx.DB, username, passwordHash, gpgKey string) error {
    err = db.CreateUser(db.Postgres, req.Username, passwordHash, mnemonic)
    if err != nil {
        http.Error(w, "failed to create user", http.StatusInternalServerError)
        return
    }


    resp := Response{
        Message:   "Account created successfully. Save your recovery phrase!",
        Encrypted: mnemonic,
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}



