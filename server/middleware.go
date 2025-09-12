package server

import (
	"context"
	"net/http"
	"strings"

	"mFrelance/auth"
)


type contextKey string

const userContextKey = contextKey("user")

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing Authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid Authorization header", http.StatusUnauthorized)
			return
		}

		claims, err := auth.ParseJWT(parts[1])
		if err != nil {
			http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Кладём claims в context
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
