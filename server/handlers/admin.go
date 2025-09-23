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
// @Summary Grant Administrative Privileges
// @Description Elevates a regular user to administrator status. Requires existing admin privileges to execute.
// @Tags administration
// @Accept json
// @Produce json
// @Param request body AdminRequest true "User ID to promote to admin"
// @Success 200 {string} string "Example: \"user is now admin\""
// @Failure 400 {string} string "Example: \"invalid request body\""
// @Failure 401 {string} string "Example: \"unauthorized\""
// @Failure 403 {string} string "Example: \"admin rights required\""
// @Failure 500 {string} string "Example: \"internal server error\""
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
// @Summary Revoke Administrative Privileges
// @Description Removes administrator status from a user, reverting them to regular user privileges.
// @Tags administration
// @Accept json
// @Produce json
// @Param request body AdminRequest true "User ID to demote from admin"
// @Success 200 {string} string "Example: \"user admin removed\""
// @Failure 400 {string} string "Example: \"invalid request body\""
// @Failure 401 {string} string "Example: \"unauthorized\""
// @Failure 403 {string} string "Example: \"admin rights required\""
// @Failure 500 {string} string "Example: \"internal server error\""
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
// @Tags administration
// @Produce json
// @Success 200 {object} map[string]interface{} "Example: {\"user_id\": 123, \"is_admin\": true}"
// @Failure 400 {string} string "Example: \"invalid request body\""
// @Failure 500 {string} string "Example: \"internal server error\""
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
// @Summary Check if current user is admin
// @Description Returns true/false if the authenticated user has admin privileges
// @Tags administration
// @Produce json
// @Success 200 {object} map[string]interface{} "Example: {\"user_id\": 123, \"is_admin\": true}"
// @Failure 401 {string} string "Example: \"user not found in context\""
// @Failure 500 {string} string "Example: \"internal server error\""
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
// @Summary Block User Account
// @Description Blocks a user account by user ID, preventing them from accessing the system
// @Tags administration
// @Accept json
// @Produce json
// @Param request body AdminRequest true "User ID to block"
// @Success 200 {string} string "Example: \"user blocked\""
// @Failure 400 {string} string "Example: \"invalid request body\""
// @Failure 500 {string} string "Example: \"internal server error\""
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
// @Summary Admin: View Transactions
// @Description Allows admin to view transactions by wallet or all transactions with pagination
// @Tags administration
// @Accept json
// @Produce json
// @Param request body AdminTransactionsRequest true "Request payload with optional wallet_id, limit, offset"
// @Success 200 {array} object "Example: [{\"id\": 123, \"from_wallet_id\": 456, \"to_wallet_id\": 789, \"to_address\": \"1ABC...\", \"task_id\": 101, \"amount\": \"0.5\", \"currency\": \"BTC\", \"confirmed\": true, \"created_at\": \"2023-12-01T10:00:00Z\"}]"
// @Failure 400 {string} string "Example: \"invalid request body\""
// @Failure 500 {string} string "Example: \"DB error\""
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
// @Summary Unblock User Account
// @Description Unblocks a previously blocked user account, restoring their access to the system
// @Tags administration
// @Accept json
// @Produce json
// @Param request body AdminRequest true "User ID to unblock"
// @Success 200 {string} string "Example: \"user unblocked\""
// @Failure 400 {string} string "Example: \"invalid request body\""
// @Failure 500 {string} string "Example: \"internal server error\""
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
// @Summary Get User Wallets
// @Description Returns all cryptocurrency wallets for a given user
// @Tags administration
// @Produce json
// @Param user_id query int true "User ID"
// @Success 200 {array} models.Wallet "Example: [{\"id\": 123, \"user_id\": 456, \"currency\": \"BTC\", \"address\": \"1ABC...\", \"balance\": \"0.5\", \"created_at\": \"2023-12-01T10:00:00Z\"}]"
// @Failure 400 {string} string "Example: \"user_id required\""
// @Failure 500 {string} string "Example: \"DB error\""
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
// @Summary Update Wallet Balance
// @Description Allows admin to manually set a new balance for a user's wallet
// @Tags administration
// @Accept json
// @Produce json
// @Param request body AdminUpdateBalanceRequest true "Wallet balance payload with user_id and balance"
// @Success 200 {string} string "Example: \"balance updated\""
// @Failure 400 {string} string "Example: \"invalid balance format\""
// @Failure 500 {string} string "Example: \"DB error\""
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

// AdminGetRandomTicketHandler godoc
// @Summary Assign Random Open Ticket
// @Description Assigns a random open support ticket to the current admin user for handling
// @Tags administration
// @Produce json
// @Success 200 {object} models.TicketDoc "Example: {\"id\": 123, \"user_id\": 456, \"admin_id\": 789, \"subject\": \"Login issue\", \"status\": \"pending\", \"created_at\": \"2023-12-01T10:00:00Z\"}"
// @Failure 400 {object} map[string]string "Example: {\"error\": \"no open tickets available\"}"
// @Failure 401 {object} map[string]string "Example: {\"error\": \"user not found in context\"}"
// @Security BearerAuth
// @Router /api/admin/getRandomTicket [get]
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

// AdminDeleteUserTasksHandler godoc
// @Summary Delete All User Tasks
// @Description Allows an admin to delete all tasks belonging to a specific user by their ID
// @Tags administration
// @Produce json
// @Param user_id query int true "User ID whose tasks to delete"
// @Success 200 {object} map[string]interface{} "Example: {\"success\": true, \"deleted\": 5}"
// @Failure 400 {object} map[string]string "Example: {\"error\": \"Invalid user_id\"}"
// @Failure 401 {object} map[string]string "Example: {\"error\": \"Unauthorized\"}"
// @Failure 403 {object} map[string]string "Example: {\"error\": \"admin rights required\"}"
// @Failure 405 {object} map[string]string "Example: {\"error\": \"Method not allowed\"}"
// @Failure 500 {object} map[string]string "Example: {\"error\": \"Failed to delete tasks\"}"
// @Router /api/admin/delete_user_tasks [post]
// @Security BearerAuth
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
