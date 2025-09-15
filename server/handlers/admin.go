package handlers

import (
    "encoding/json"
    "math/big"
    "net/http"
    "strconv"

    "mFrelance/db"
    "mFrelance/models"
    "mFrelance/server"
)

type AdminRequest struct {
    UserID int64 `json:"user_id"`
}

func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        claims := server.GetUserFromContext(r)
        if claims == nil {
            http.Error(w, "user not found (RequireAdmin)", http.StatusUnauthorized)
            return
        }
        isAdmin, err := db.IsAdmin(db.Postgres, claims.UserID)
        if err != nil || !isAdmin {
            http.Error(w, "admin rights required", http.StatusForbidden)
            return
        }
        next(w, r)
    }
}

func MakeAdminHandler(w http.ResponseWriter, r *http.Request) {
    var req AdminRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }
    if err := db.MakeAdmin(db.Postgres, req.UserID); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("user is now admin"))
}

func RemoveAdminHandler(w http.ResponseWriter, r *http.Request) {
    var req AdminRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }
    if err := db.RemoveAdmin(db.Postgres, req.UserID); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("user admin removed"))
}

func IsAdminHandler(w http.ResponseWriter, r *http.Request) {
    var req AdminRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }
    isAdmin, err := db.IsAdmin(db.Postgres, req.UserID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(map[string]any{
        "user_id":  req.UserID,
        "is_admin": isAdmin,
    })
}

func BlockUserHandler(w http.ResponseWriter, r *http.Request) {
    var req AdminRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }
    if err := db.BlockUser(db.Postgres, req.UserID); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("user blocked"))
}

type AdminTransactionsRequest struct {
    WalletID int `json:"wallet_id,omitempty"`
    Limit    int `json:"limit,omitempty"`
    Offset   int `json:"offset,omitempty"`
}

func AdminTransactionsHandler(w http.ResponseWriter, r *http.Request) {
    var req AdminTransactionsRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }
    limit := req.Limit
    offset := req.Offset
    if limit <= 0 {
        limit = 50
    }
    if offset < 0 {
        offset = 0
    }
    var txs []*models.Transaction
    var err error
    if req.WalletID > 0 {
        txs, err = models.GetTransactionsByWallet(int64(req.WalletID), limit, offset)
    } else {
        txs, err = models.GetTransactions(limit, offset)
    }
    if err != nil {
        http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(txs)
}

func UnblockUserHandler(w http.ResponseWriter, r *http.Request) {
    var req AdminRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }
    if err := db.UnblockUser(db.Postgres, req.UserID); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("user unblocked"))
}

func AdminWalletsHandler(w http.ResponseWriter, r *http.Request) {
    userIDStr := r.URL.Query().Get("user_id")
    if userIDStr == "" {
        http.Error(w, "user_id required", http.StatusBadRequest)
        return
    }
    userID, err := strconv.ParseInt(userIDStr, 10, 64)
    if err != nil {
        http.Error(w, "invalid user_id", http.StatusBadRequest)
        return
    }
    wallets, err := models.GetWalletsByUser(userID)
    if err != nil {
        http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(wallets)
}

type AdminUpdateBalanceRequest struct {
    UserID  int64  `json:"user_id"`
    Balance string `json:"balance"`
}

func AdminUpdateBalanceHandler(w http.ResponseWriter, r *http.Request) {
    var req AdminUpdateBalanceRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }
    newBalance, ok := new(big.Float).SetString(req.Balance)
    if !ok {
        http.Error(w, "invalid balance format", http.StatusBadRequest)
        return
    }
    _, err := db.Postgres.Exec(`UPDATE wallets SET balance=$1 WHERE user_id=$2`, newBalance.Text('f', 12), req.UserID)
    if err != nil {
        http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("balance updated"))
}

// AdminGetRandomTicketHandler assigns a random open ticket to the current admin
func AdminGetRandomTicketHandler(w http.ResponseWriter, r *http.Request) {
    claims := server.GetUserFromContext(r)
    if claims == nil {
        server.WriteErrorJSON(w, "user not found in context", http.StatusUnauthorized)
        return
    }

    ticket, err := models.GetRandomOpenTicket()
    if err != nil {
        server.WriteErrorJSON(w, err.Error(), http.StatusBadRequest)
        return
    }
    if err := models.AssignTicketAdmin(ticket.ID, claims.UserID); err != nil {
        server.WriteErrorJSON(w, err.Error(), http.StatusBadRequest)
        return
    }
    if err := models.PendingTicket(ticket.ID); err != nil {
        server.WriteErrorJSON(w, err.Error(), http.StatusBadRequest)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(ticket)
}


