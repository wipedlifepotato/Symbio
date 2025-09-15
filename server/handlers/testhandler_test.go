package handlers_test

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "mFrelance/auth"
    "mFrelance/server"
    "mFrelance/server/handlers"
)

func TestTestHandler_Unauthorized(t *testing.T) {
    h := server.AuthMiddleware(http.HandlerFunc(handlers.TestHandler))

    req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
    rr := httptest.NewRecorder()
    h.ServeHTTP(rr, req)

    if rr.Code != http.StatusUnauthorized {
        t.Fatalf("want 401, got %d body=%s", rr.Code, rr.Body.String())
    }
}

func TestTestHandler_Authorized(t *testing.T) {
    h := server.AuthMiddleware(http.HandlerFunc(handlers.TestHandler))

    req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
    token, err := auth.GenerateJWT(42, "tester")
    if err != nil { t.Fatalf("jwt: %v", err) }
    req.Header.Set("Authorization", "Bearer "+token)

    rr := httptest.NewRecorder()
    h.ServeHTTP(rr, req)

    if rr.Code != http.StatusOK {
        t.Fatalf("want 200, got %d body=%s", rr.Code, rr.Body.String())
    }
}


