package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/steambap/captcha"

	"mFrelance/auth"
	"mFrelance/db"
	"mFrelance/server"
)
import "unicode/utf8"

type RegisterRequest struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	CaptchaID     string `json:"captcha_id"`
	CaptchaAnswer string `json:"captcha_answer"`
}

type Response struct {
	Message   string `json:"message"`
	Encrypted string `json:"encrypted,omitempty"`
}

// HelloHandler godoc
// @Summary Health/hello
// @Description Simple hello endpoint
// @Tags auth
// @Produce json
// @Success 200 {object} Response
// @Router /hello [get]
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

// CaptchaHandler godoc
// @Summary Get captcha image
// @Description Returns a captcha image and X-Captcha-ID header
// @Tags auth
// @Produce png
// @Success 200 "image/png"
// @Header 200 {string} X-Captcha-ID "Captcha ID"
// @Router /captcha [get]
func CaptchaHandler(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {
	data, _ := captcha.New(150, 50)
	id := strconv.Itoa(int(time.Now().UnixNano()))
	rdb.Set(ctx, "captcha:"+id, data.Text, 5*time.Minute)

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("X-Captcha-ID", id)
	data.WriteImage(w)
}

// VerifyHandler godoc
// @Summary Verify captcha
// @Description Verifies provided captcha answer
// @Tags auth
// @Param id query string true "Captcha ID"
// @Param answer query string true "Captcha answer"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} map[string]string
// @Router /verify [get]
func VerifyHandler(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {
	id := r.URL.Query().Get("id")
	answer := r.URL.Query().Get("answer")

	stored, err := rdb.Get(ctx, "captcha:"+id).Result()
	if err != nil {
		server.WriteErrorJSON(w, "Captcha expired", http.StatusBadRequest)
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
// @Success 200 {object} Response
// @Failure 400 {object} map[string]string
// @Router /register [post]
func RegisterHandler(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {
	log.Println("[RegisterHandler] Register Handler")
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		server.WriteErrorJSON(w, "invalid json", http.StatusBadRequest)
		log.Println("[RegisterHandler] invalid json")
		return
	}
	if utf8.RuneCountInString(req.Password) > 128 {
		server.WriteErrorJSON(w, "password too long", http.StatusBadRequest)
		log.Println("[RegisterHandler] password too long")
		return
	}
	log.Println(req.Password)
	if utf8.RuneCountInString(req.Password) < 6 {
		server.WriteErrorJSON(w, "password too small", http.StatusBadRequest)
		log.Println("[RegisterHandler] password too small")
		return
	}
	if utf8.RuneCountInString(req.Username) < 2 {
		server.WriteErrorJSON(w, "Username too small", http.StatusBadRequest)
		log.Println("[RegisterHandler] username too small")
		return
	}
	if utf8.RuneCountInString(req.Username) > 128 {
		server.WriteErrorJSON(w, "Username too long", http.StatusBadRequest)
		log.Println("[RegisterHandler] username too long")
		return
	}
	storedCaptcha, err := rdb.Get(ctx, "captcha:"+req.CaptchaID).Result()
	if err != nil || storedCaptcha != req.CaptchaAnswer {
		server.WriteErrorJSON(w, "invalid captcha", http.StatusBadRequest)
		log.Println("[RegisterHandler] invalid captcha")
		return
	}
	rdb.Del(ctx, "captcha:"+req.CaptchaID)

	mnemonic := server.GenerateMnemonic()
	passwordHash := server.HashPassword(req.Password)

	err = db.CreateUser(db.Postgres, req.Username, passwordHash, mnemonic)
	if err != nil {
		log.Println("[RegisterHandler] failed to create user")
		server.WriteErrorJSON(w, "failed to create user, maybe user exists", http.StatusInternalServerError)
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

// RestoreHandler godoc
// @Summary Restore user account
// @Description Restore account by mnemonic and set new password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RestoreRequest true "Restore payload"
// @Success 200 {object} Response
// @Failure 400 {object} map[string]string
// @Router /restoreuser [post]
func RestoreHandler(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {
	var req RestoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		server.WriteErrorJSON(w, "invalid json", http.StatusBadRequest)
		return
	}
	if utf8.RuneCountInString(req.NewPassword) > 128 {
		server.WriteErrorJSON(w, "password too long", http.StatusBadRequest)
		log.Println("[RegisterHandler] password too long")
		return
	}
	log.Println(req.NewPassword)
	if utf8.RuneCountInString(req.NewPassword) < 6 {
		server.WriteErrorJSON(w, "password too small", http.StatusBadRequest)
		log.Println("[RegisterHandler] password too small")
		return
	}
	storedCaptcha, err := rdb.Get(ctx, "captcha:"+req.CaptchaID).Result()
	if err != nil || storedCaptcha != req.CaptchaAnswer {
		server.WriteErrorJSON(w, "invalid captcha", http.StatusBadRequest)
		return
	}
	rdb.Del(ctx, "captcha:"+req.CaptchaID)

	userID, username, err := db.RestoreUser(db.Postgres, req.Username, req.Mnemonic)
	if err != nil || userID == 0 || username == "" {
		server.WriteErrorJSON(w, "failed to found user", http.StatusInternalServerError)
		return
	}
	passwordHash := server.HashPassword(req.NewPassword)
	err = db.ChangeUserPassword(db.Postgres, req.Username, passwordHash)
	if err != nil {
		server.WriteErrorJSON(w, "Failed to change user password", http.StatusInternalServerError)
		return
	}
	token, err := auth.GenerateJWT(userID, username)
	if err != nil {
		server.WriteErrorJSON(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}
	resp := Response{
		Message:   "Account restored successfully.",
		Encrypted: token,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type AuthRequest struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	CaptchaID     string `json:"captcha_id"`
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
// @Success 200 {object} AuthResponse
// @Failure 401 {object} map[string]string
// @Router /auth [post]
func AuthHandler(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		server.WriteErrorJSON(w, "invalid json", http.StatusBadRequest)
		return
	}
	log.Print("[AuthHandler] Test captcha")
	storedCaptcha, err := rdb.Get(ctx, "captcha:"+req.CaptchaID).Result()
	if err != nil || storedCaptcha != req.CaptchaAnswer {
		server.WriteErrorJSON(w, "invalid captcha", http.StatusBadRequest)
		return
	}
	rdb.Del(ctx, "captcha:"+req.CaptchaID)

	log.Print("[AuthHandler] Get User by Username")
	userID, passwordHash, err := db.GetUserByUsername(db.Postgres, req.Username)
	if err != nil {
		server.WriteErrorJSON(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if userID == 0 {
		server.WriteErrorJSON(w, "invalid username or password", http.StatusUnauthorized)
		return
	}
	log.Print("[AuthHandler] Check Password")
	if !server.VerifyPassword(req.Password, passwordHash) {
		server.WriteErrorJSON(w, "invalid username or password", http.StatusUnauthorized)
		return
	}
	log.Print("[AuthHandler] Generate JWT")
	token, err := auth.GenerateJWT(userID, req.Username)
	if err != nil {
		server.WriteErrorJSON(w, "failed to generate token", http.StatusInternalServerError)
		return
	}
	resp := AuthResponse{Message: "Authenticated successfully", Token: token}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
