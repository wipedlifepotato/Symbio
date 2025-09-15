package handlers_test

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "strings"

    "mFrelance/server/handlers"
    "mFrelance/server/testutil"
)

func TestHelloHandler(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/hello", nil)
    rr := httptest.NewRecorder()

    handlers.HelloHandler(rr, req)

    if rr.Code != http.StatusOK {
        t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
    }
}

func TestCaptchaHandler_SaveToRedis(t *testing.T) {
    mr, rdb := testutil.NewMiniRedis(t)
    defer mr.Close()

    req := httptest.NewRequest(http.MethodGet, "/captcha", nil)
    rr := httptest.NewRecorder()

    handlers.CaptchaHandler(rr, req, rdb)

    if rr.Code != http.StatusOK {
        t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
    }
    id := rr.Header().Get("X-Captcha-ID")
    if id == "" {
        t.Fatalf("missing X-Captcha-ID header")
    }
    if !mr.Exists("captcha:" + id) {
        t.Fatalf("captcha not saved in redis")
    }
}

func TestVerifyHandler_OKAndFail(t *testing.T) {
    mr, rdb := testutil.NewMiniRedis(t)
    defer mr.Close()

    // arrange captcha value
    id := "12345"
    mr.Set("captcha:"+id, "abcd")

    // ok case
    reqOK := httptest.NewRequest(http.MethodGet, "/verify?id="+id+"&answer=abcd", nil)
    rrOK := httptest.NewRecorder()
    handlers.VerifyHandler(rrOK, reqOK, rdb)
    if rrOK.Code != http.StatusOK {
        t.Fatalf("verify ok status=%d", rrOK.Code)
    }

    // fail case (wrong answer)
    mr.Set("captcha:"+id, "abcd")
    reqBad := httptest.NewRequest(http.MethodGet, "/verify?id="+id+"&answer=zzz", nil)
    rrBad := httptest.NewRecorder()
    handlers.VerifyHandler(rrBad, reqBad, rdb)
    if rrBad.Code != http.StatusOK {
        t.Fatalf("verify bad status=%d", rrBad.Code)
    }
}

func TestAuthHandler_InvalidJSON(t *testing.T) {
    _, rdb := testutil.NewMiniRedis(t)

    req := httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader("not-json"))
    rr := httptest.NewRecorder()
    handlers.AuthHandler(rr, req, rdb)
    if rr.Code != http.StatusBadRequest {
        t.Fatalf("want 400, got %d", rr.Code)
    }
}


