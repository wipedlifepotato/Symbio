package server

import (

	"encoding/json"
    //"log"
    //"mFrelance/db"
    "net/http"
    //"strconv"
    //"mFrelance/models"

	)

func TestHandler(w http.ResponseWriter, r *http.Request) {
	claims := GetUserFromContext(r)
	if claims == nil {
		writeErrorJSON(w, "user not found in context", http.StatusUnauthorized)
		return
	}

	resp := map[string]any{
		"user_id": claims.UserID,
		"username": claims.Username,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}



