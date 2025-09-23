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

// AdminRequest payload
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

// MakeAdminHandler godoc
// @Summary Grant admin rights
// @Description Makes a user admin by userID
// @Tags admin
// @Accept json
// @Produce json
// @Param request body AdminRequest true "UserID payload"
// @Success 200 {string} string "user is now admin"
// @Failure 400 {string} string "invalid request body"
// @Failure 401 {string} string "unauthorized"
// @Failure 403 {string} string "admin rights required"
// @Failure 500 {string} string "internal server error"
// @Security BearerAuth
// @Router /api/admin/make [post]
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

// RemoveAdminHandler godoc
// @Summary Revoke admin rights
// @Description Removes admin status from a user
// @Tags admin
// @Accept json
// @Produce json
// @Param request body AdminRequest true "UserID payload"
// @Success 200 {string} string "user admin removed"
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Security BearerAuth
// @Router /api/admin/remove [post]
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

// IsAdminHandler godoc
// @Summary Check if user is admin
// @Description Returns true/false if current user has admin privileges
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Router /api/admin/check [get]
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

// IsIAdminHandler godoc
// @Summary Check if user is admin
// @Description Returns true/false if current user has admin privileges
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Router /api/admin/IIsAdmin [get]
func IsIAdminHandler(w http.ResponseWriter, r *http.Request) {
	claims := server.GetUserFromContext(r)
	if claims == nil {
		server.WriteErrorJSON(w, "user not found in context", http.StatusUnauthorized)
		return
	}
	isAdmin, err := db.IsAdmin(db.Postgres, claims.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]any{
		"user_id":  claims.UserID,
		"is_admin": isAdmin,
	})
}

// BlockUserHandler godoc
// @Summary Block user
// @Description Blocks a user by userID
// @Tags admin
// @Accept json
// @Produce json
// @Param request body AdminRequest true "UserID payload"
// @Success 200 {string} string "user blocked"
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Security BearerAuth
// @Router /api/admin/block [post]
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

// AdminTransactionsHandler godoc
// @Summary Admin: View transactions
// @Description Allows admin to view transactions by wallet or all transactions with pagination
// @Tags admin
// @Accept json
// @Produce json
// @Param request body AdminTransactionsRequest true "Request payload"
// @Success 200 {array} object
// @Success 200 {array} object "id:int, from_wallet_id:int, to_wallet_id:int, to_address:string, task_id:int, amount:string, currency:string, confirmed:bool, created_at:string"
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Security BearerAuth
// @Router /api/admin/transactions [post]
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
		txs, err = models.GetTransactionsByWallet(db.Postgres, int64(req.WalletID), limit, offset)
	} else {
		txs, err = models.GetTransactions(db.Postgres, limit, offset)
	}
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(txs)
}

// UnblockUserHandler godoc
// @Summary Unblock user
// @Description Unblocks a user by userID
// @Tags admin
// @Accept json
// @Produce json
// @Param request body AdminRequest true "UserID payload"
// @Success 200 {string} string "user unblocked"
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Security BearerAuth
// @Router /api/admin/unblock [post]
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

// AdminWalletsHandler godoc
// @Summary Get user wallets
// @Description Returns all wallets for a given user
// @Tags admin
// @Accept json
// @Produce json
// @Param user_id query int true "User ID"
// @Success 200 {array} models.Wallet
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Security BearerAuth
// @Router /api/admin/wallets [get]
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
	wallets, err := models.GetWalletsByUser(db.Postgres, userID)
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

// AdminUpdateBalanceHandler godoc
// @Summary Update wallet balance
// @Description Allows admin to set a new balance for a wallet
// @Tags admin
// @Accept json
// @Produce json
// @Param request body AdminUpdateBalanceRequest true "Wallet balance payload"
// @Success 200 {string} string "balance updated"
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Security BearerAuth
// @Router /api/admin/update_balance [post]
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
// AdminGetRandomTicketHandler godoc
// @Summary Get random opened ticket (admin)
// @Description Set ticket to admin (random)
// @Tags admin
// @Produce json
// @Success 200 {object} models.TicketDoc
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/admin/getRandomTicket [get]
// @Security bearerAuth
func AdminGetRandomTicketHandler(w http.ResponseWriter, r *http.Request) {
	claims := server.GetUserFromContext(r)
	if claims == nil {
		server.WriteErrorJSON(w, "user not found in context", http.StatusUnauthorized)
		return
	}

	ticket, err := models.GetRandomOpenTicket(db.Postgres)
	if err != nil {
		server.WriteErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := models.AssignTicketAdmin(db.Postgres, ticket.ID, claims.UserID); err != nil {
		server.WriteErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := models.PendingTicket(db.Postgres, ticket.ID); err != nil {
		server.WriteErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)
}
// AdminAddUserToChatRoom godoc
// @Summary Add user to chat room
// @Description Allows an admin to add a user to an existing chat room
// @Tags chats
// @Produce json
// @Param chat_id query int true "Chat room ID"
// @Param user_id query int true "User ID to add"
// @Success 200 {object} map[string]string "Result message"
// @Failure 400 {object} map[string]string "Bad chat_id or user_id"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/chats/add_user [post]
// @Security BearerAuth

func AdminAddUserToChatRoom(w http.ResponseWriter, r *http.Request) {
	claims := server.GetUserFromContext(r)
	if claims == nil {
		server.WriteErrorJSON(w, "user not found in context", http.StatusUnauthorized)
		return
	}
	chatID := r.URL.Query().Get("chat_id")
	userId := r.URL.Query().Get("user_id")
	chatIDInt, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		server.WriteErrorJSON(w, "bad chat id", http.StatusBadRequest)

	}
	userIDInt, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		server.WriteErrorJSON(w, "bad user id", http.StatusBadRequest)

	}
	err = db.AddUserToChatRoom(db.Postgres, chatIDInt, userIDInt)
	if err != nil {
		server.WriteErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"res\":\"user added to chat room\"}"))
}
// DeleteChatRoom godoc
// @Summary Delete chat room
// @Description Allows an admin to delete a chat room by ID
// @Tags chats
// @Produce json
// @Param chat_id query int true "Chat room ID"
// @Success 200 {object} map[string]string "Result message"
// @Failure 400 {object} map[string]string "Bad chat_id"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/chats/delete [delete]
// @Security BearerAuth
func DeleteChatRoom(w http.ResponseWriter, r *http.Request) {
	claims := server.GetUserFromContext(r)
	if claims == nil {
		server.WriteErrorJSON(w, "user not found in context", http.StatusUnauthorized)
		return
	}
	chatID := r.URL.Query().Get("chat_id")
	chatIDInt, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		server.WriteErrorJSON(w, "bad chat id", http.StatusBadRequest)

	}
	err = db.DeleteChatRoom(db.Postgres, chatIDInt)
	if err != nil {
		server.WriteErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"res\":\"chat room deleted\"}"))
}

// AdminDeleteUserTasksHandler deletes all tasks of a user (admin only)
func AdminDeleteUserTasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	claims := server.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	isAdmin, err := db.IsAdmin(db.Postgres, claims.UserID)
	if err != nil || !isAdmin {
		http.Error(w, "admin rights required", http.StatusForbidden)
		return
	}
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}
	n, err := db.DeleteTasksByUserID(db.Postgres, userID)
	if err != nil {
		http.Error(w, "Failed to delete tasks", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"deleted": n,
	})
}
