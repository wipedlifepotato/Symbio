package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
	"io/ioutil"
	"github.com/go-redis/redis/v8"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/rand"
	"strconv"

	"mFrelance/auth"
	"mFrelance/config"
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

func getClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}
	ip := r.RemoteAddr
	if strings.Contains(ip, ":") {
		ip, _, _ = strings.Cut(ip, ":")
	}
	return ip
}

func resetAllCaptchas(rdb *redis.Client) {
	// Find all captcha keys and delete them
	keys, _ := rdb.Keys(ctx, "captcha:*").Result()
	if len(keys) > 0 {
		rdb.Del(ctx, keys...)
	}
}

func generateCaptchaText() string {
	rand.Seed(time.Now().UnixNano())
	return strconv.Itoa(1000 + rand.Intn(9000)) // 4-digit number
}

func generateCaptchaImage(text string) image.Image {
	width, height := 200, 60
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	for i := 0; i < 100; i++ {
		x := rand.Intn(width)
		y := rand.Intn(height)
		img.Set(x, y, color.Black)
	}

	fontBytes, err := ioutil.ReadFile(config.AppConfig.CaptchaFontPath)
	if err != nil {
		log.Fatal(err)
	}
	ttfFont, err := opentype.Parse(fontBytes)
	if err != nil {
		log.Fatal(err)
	}

	face, err := opentype.NewFace(ttfFont, &opentype.FaceOptions{
		Size:    32,          // размер шрифта (можно увеличить)
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.Black),
		Face: face,
	}

	textWidth := d.MeasureString(text).Ceil()
	x := (width - textWidth) / 2
	y := (height + int(face.Metrics().Ascent.Ceil()) - int(face.Metrics().Descent.Ceil())) / 2

	d.Dot = fixed.Point26_6{
		X: fixed.I(x),
		Y: fixed.I(y),
	}

	d.DrawString(text)

	return img
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
	if !config.AppConfig.CaptchaEnabled {
		server.WriteErrorJSON(w, "Captcha is disabled", http.StatusServiceUnavailable)
		return
	}

	ip := getClientIP(r)

	minuteKey := "captcha:count:" + ip + ":minute"
	hourKey := "captcha:count:" + ip + ":hour"

	minuteCount, _ := rdb.Incr(ctx, minuteKey).Result()
	if minuteCount == 1 {
		rdb.Expire(ctx, minuteKey, time.Minute)
	}

	hourCount, _ := rdb.Incr(ctx, hourKey).Result()
	if hourCount == 1 {
		rdb.Expire(ctx, hourKey, time.Hour)
	}

	if minuteCount > int64(config.AppConfig.CaptchaRateLimitPerMinute) || hourCount > int64(config.AppConfig.CaptchaRateLimitPerHour) {
		// If hour limit exceeded, reset all captcha keys to prevent DoS
		if hourCount > int64(config.AppConfig.CaptchaRateLimitPerHour) {
			resetAllCaptchas(rdb)
		}
		server.WriteErrorJSON(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
		return
	}

	text := generateCaptchaText()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	rdb.Set(ctx, "captcha:"+id, text, 5*time.Minute)

	img := generateCaptchaImage(text)
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("X-Captcha-ID", id)
	png.Encode(w, img)
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

// CaptchaStatusHandler godoc
// @Summary Check captcha status
// @Description Returns whether captcha is enabled
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]bool
// @Router /captcha/status [get]
func CaptchaStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"enabled": config.AppConfig.CaptchaEnabled})
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
