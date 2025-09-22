package handlers

import (
	"encoding/json"
	"mFrelance/db"
	"mFrelance/models"
	"mFrelance/server"
	"math/big"
	"net/http"
	"strconv"
)
// GetOpenDisputesHandler godoc
// @Summary Get all open disputes
// @Description Returns a list of disputes with status "open"
// @Tags disputes
// @Produce json
// @Success 200 {object} map[string]interface{} "success flag and disputes list"
// @Failure 500 {string} string "Failed to get disputes"
// @Router /api/disputes/open [get]
// @Security BearerAuth
func GetOpenDisputesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		disputes, err := db.GetOpenDisputes()
		if err != nil {
			http.Error(w, "Failed to get disputes", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"disputes": disputes,
		})
	}
}

type AssignDisputeRequest struct {
    DisputeID int64 `json:"dispute_id"`
}
// @Summary Assign dispute
// @Description Assign a dispute to the current admin
// @Tags disputes
// @Accept json
// @Produce json
// @Param body body AssignDisputeRequest true "Dispute ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/disputes/assign [post]
func AssignDisputeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			DisputeID int64 `json:"dispute_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		claims := server.GetUserFromContext(r)
		if claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		adminID := claims.UserID

		_, err := db.GetDisputeByID(req.DisputeID)
		if err != nil {
			http.Error(w, "Dispute not found", http.StatusNotFound)
			return
		}

		if err := db.AssignDisputeToAdmin(req.DisputeID, adminID); err != nil {
			http.Error(w, "Failed to assign dispute", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Dispute assigned successfully",
		})
	}
}


type ResolveDisputeRequest struct {
    DisputeID  int64  `json:"dispute_id"`
    Resolution string `json:"resolution"` // "client_won" или "freelancer_won"
}
// ResolveDisputeHandler godoc
// @Summary Resolve dispute
// @Description Allows an assigned admin to resolve a dispute and release funds from escrow
// @Tags disputes
// @Accept json
// @Produce json
// @Param body body ResolveDisputeRequest true "Resolution payload"
// @Success 200 {object} map[string]interface{} "Success flag and message"
// @Failure 400 {object} map[string]interface{} "Invalid JSON or invalid resolution"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Dispute not assigned to you"
// @Failure 404 {object} map[string]interface{} "Dispute or task not found"
// @Failure 500 {object} map[string]interface{} "Failed to resolve dispute"
// @Router /api/disputes/resolve [post]
// @Security BearerAuth
func ResolveDisputeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			DisputeID  int64  `json:"dispute_id"`
			Resolution string `json:"resolution"` // "client_won" или "freelancer_won"
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		claims := server.GetUserFromContext(r)
		if claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		adminID := claims.UserID

		dispute, err := db.GetDisputeByID(req.DisputeID)
		if err != nil {
			http.Error(w, "Dispute not found", http.StatusNotFound)
			return
		}

		if dispute.AssignedAdmin == nil || *dispute.AssignedAdmin != adminID {
			http.Error(w, "Dispute not assigned to you", http.StatusForbidden)
			return
		}

		if dispute.Status != "open" {
			http.Error(w, "Dispute is not open", http.StatusBadRequest)
			return
		}

		task, err := db.GetTask(db.Postgres, dispute.TaskID)
		if err != nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		escrow, err := db.GetEscrowBalanceByTaskID(task.ID)
		if err != nil {
			http.Error(w, "Escrow balance not found", http.StatusNotFound)
			return
		}

		amount := big.NewFloat(escrow.Amount)

		var walletUserID int64
		var newEscrowStatus string
		if req.Resolution == "client_won" {
			walletUserID = task.ClientID
			newEscrowStatus = "refunded"
		} else if req.Resolution == "freelancer_won" {
			walletUserID = escrow.FreelancerID
			newEscrowStatus = "released"
		} else {
			http.Error(w, "Invalid resolution", http.StatusBadRequest)
			return
		}

		userWallet, err := models.GetWalletByUserAndCurrency(db.Postgres, walletUserID, task.Currency)
		if err != nil {
			http.Error(w, "Failed to get user wallet: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := userWallet.AddBalance(db.Postgres, amount); err != nil {
			http.Error(w, "Failed to credit user: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := db.UpdateEscrowBalanceStatus(task.ID, newEscrowStatus); err != nil {
			http.Error(w, "Failed to update escrow status", http.StatusInternalServerError)
			return
		}

		if err := db.UpdateDisputeStatus(req.DisputeID, "resolved", &req.Resolution); err != nil {
			http.Error(w, "Failed to update dispute status", http.StatusInternalServerError)
			return
		}

		if err := db.UpdateTaskStatus(db.Postgres, task.ID, "completed"); err != nil {
			http.Error(w, "Failed to update task status", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Dispute resolved and funds transferred successfully",
		})
	}
}
// GetDisputeDetailsHandler godoc
// @Summary Get dispute details
// @Description Returns dispute info, related task, escrow balance and messages
// @Tags disputes
// @Produce json
// @Param id query int true "Dispute ID"
// @Success 200 {object} map[string]interface{} "success flag and dispute details"
// @Failure 400 {string} string "Invalid dispute ID"
// @Failure 404 {string} string "Dispute, task or escrow not found"
// @Failure 500 {string} string "Failed to get messages"
// @Router /api/disputes/details [get]
// @Security BearerAuth
func GetDisputeDetailsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		disputeIDStr := r.URL.Query().Get("id")
		disputeID, err := strconv.ParseInt(disputeIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid dispute ID", http.StatusBadRequest)
			return
		}

		dispute, err := db.GetDisputeByID(disputeID)
		if err != nil {
			http.Error(w, "Dispute not found", http.StatusNotFound)
			return
		}

		task, err := db.GetTask(db.Postgres, dispute.TaskID)
		if err != nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		escrow, err := db.GetEscrowBalanceByTaskID(dispute.TaskID)
		if err != nil {
			http.Error(w, "Escrow balance not found", http.StatusNotFound)
			return
		}

		messages, err := db.GetDisputeMessages(disputeID)
		if err != nil {
			http.Error(w, "Failed to get dispute messages", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"dispute":  dispute,
			"task":     task,
			"escrow":   escrow,
			"messages": messages,
		})
	}
}
