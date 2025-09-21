package server

import (
	"bytes"
	"context"
	"io"
	"log"
	"mFrelance/auth"
	"net/http"
	"strings"
)

type contextKey string

const userContextKey = contextKey("user")

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		log.Printf("[AuthMiddleware] Request body: %s", string(body))

		r.Body = io.NopCloser(bytes.NewBuffer(body))
		log.Printf("[AuthMiddleware] Processing request to %s", r.URL.Path)

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Printf("[AuthMiddleware] Missing Authorization header")
			http.Error(w, "missing Authorization header", http.StatusUnauthorized)
			return
		}
		for name, values := range r.Header {
			for _, v := range values {
				log.Printf("[AuthMiddleware] Header: %s=%s", name, v)
			}
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Printf("[AuthMiddleware] Invalid Authorization header format: %s", authHeader)
			http.Error(w, "invalid Authorization header", http.StatusUnauthorized)
			return
		}

		log.Printf("[AuthMiddleware] Parsing JWT token")
		claims, err := auth.ParseJWT(parts[1])
		if err != nil {
			log.Printf("[AuthMiddleware] JWT parse error: %v", err)
			http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		log.Printf("[AuthMiddleware] User authenticated: ID=%d, Username=%s", claims.UserID, claims.Username)

		ctx := context.WithValue(r.Context(), userContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserFromContext(r *http.Request) *auth.Claims {
	claims, ok := r.Context().Value(userContextKey).(*auth.Claims)
	if !ok {
		return nil
	}
	return claims
}
