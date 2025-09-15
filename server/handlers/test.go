package handlers

import (
    "encoding/json"
    "net/http"
    "mFrelance/server"
)

func writeErrorJSON(w http.ResponseWriter, msg string, code int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func TestHandler(w http.ResponseWriter, r *http.Request) {
    claims := server.GetUserFromContext(r)
    if claims == nil {
        writeErrorJSON(w, "user not found in context", http.StatusUnauthorized)
        return
    }

    resp := map[string]any{
        "user_id":  claims.UserID,
        "username": claims.Username,
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}


