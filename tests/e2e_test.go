//go:build e2e

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

type jsonMap = map[string]any

func baseURL() string {
	if v := os.Getenv("BASE_URL"); v != "" {
		return v
	}
	return "http://127.0.0.1:8080"
}

func httpGet(t *testing.T, path string, headers map[string]string) (*http.Response, []byte) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, baseURL()+path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	return resp, body
}

func httpPostJSON(t *testing.T, path string, payload any, headers map[string]string) (*http.Response, []byte) {
	t.Helper()
	var buf bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&buf).Encode(payload); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}
	req, _ := http.NewRequest(http.MethodPost, baseURL()+path, &buf)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	return resp, body
}

func rdbFromEnv(t *testing.T) *redis.Client {
	t.Helper()
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR not set; skipping captcha-dependent tests")
	}
	return redis.NewClient(&redis.Options{Addr: addr})
}

func randomUsername() string {
	rand.Seed(time.Now().UnixNano())
	return "user_" + strconv.Itoa(rand.Intn(1_000_000_000))
}

func Test_Register_Auth_Wallet_E2E(t *testing.T) {
	// 1) GET /captcha → X-Captcha-ID
	resp, _ := httpGet(t, "/captcha", nil)
	if resp.StatusCode != http.StatusOK {
		t.Skipf("captcha not available: %d", resp.StatusCode)
	}
	captchaID := resp.Header.Get("X-Captcha-ID")
	if captchaID == "" {
		t.Fatalf("no X-Captcha-ID header")
	}

	// 2) Override captcha answer directly in Redis (needs REDIS_ADDR)
	rdb := rdbFromEnv(t)
	if err := rdb.Set(context.Background(), "captcha:"+captchaID, "abcd", 2*time.Minute).Err(); err != nil {
		t.Fatalf("redis set captcha: %v", err)
	}

	// 3) POST /register
	username := randomUsername()
	regPayload := jsonMap{
		"username":       username,
		"password":       "pass12345",
		"captcha_id":     captchaID,
		"captcha_answer": "abcd",
	}
	resp, body := httpPostJSON(t, "/register", regPayload, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("register status=%d body=%s", resp.StatusCode, string(body))
	}

	// 4) Prepare captcha for auth
	resp, _ = httpGet(t, "/captcha", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("captcha 2 status=%d", resp.StatusCode)
	}
	captchaID2 := resp.Header.Get("X-Captcha-ID")
	if captchaID2 == "" {
		t.Fatalf("no X-Captcha-ID for auth")
	}
	if err := rdb.Set(context.Background(), "captcha:"+captchaID2, "efgh", 2*time.Minute).Err(); err != nil {
		t.Fatalf("redis set captcha2: %v", err)
	}

	// 5) POST /auth
	authPayload := jsonMap{
		"username":       username,
		"password":       "pass12345",
		"captcha_id":     captchaID2,
		"captcha_answer": "efgh",
	}
	resp, body = httpPostJSON(t, "/auth", authPayload, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("auth status=%d body=%s", resp.StatusCode, string(body))
	}
	var authResp struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(body, &authResp)
	if authResp.Token == "" {
		t.Fatalf("empty token")
	}

	// 6) GET /api/wallet?currency=BTC
	headers := map[string]string{"Authorization": "Bearer " + authResp.Token}
	resp, body = httpGet(t, "/api/wallet?currency=BTC", headers)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("wallet status=%d body=%s", resp.StatusCode, string(body))
	}
}
