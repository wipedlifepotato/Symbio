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
// @Summary Retrieve Open Disputes
// @Description Returns a paginated list of all unresolved disputes requiring admin attention. Used for dispute management dashboard.
// @Tags dispute-management
// @Produce json
// @Success 200 {object} map[string]interface{} "success: true, disputes: array of open dispute objects"
// @Failure 401 {string} string "Authentication required"
// @Failure 403 {string} string "Insufficient permissions"
// @Failure 500 {string} string "Database error retrieving disputes"
// @Router /api/admin/disputes [get]
// @Security BearerAuth
func GetOpenDisputesHandler() http.HandlerFunc {
	return server.RequirePermission(server.PermDisputeManage)(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse query parameters for filtering
		idStr := r.URL.Query().Get("id")
		taskIDStr := r.URL.Query().Get("task_id")
		status := r.URL.Query().Get("status")

		var disputes []*models.Dispute
		var err error

		if idStr != "" {
			// Filter by specific dispute ID
			id, parseErr := strconv.ParseInt(idStr, 10, 64)
			if parseErr != nil {
				http.Error(w, "Invalid dispute ID", http.StatusBadRequest)
				return
			}
			dispute, getErr := db.GetDisputeByID(id)
			if getErr != nil {
				http.Error(w, "Dispute not found", http.StatusNotFound)
				return
			}
			disputes = []*models.Dispute{dispute}
		} else if taskIDStr != "" {
			// Filter by task ID
			taskID, parseErr := strconv.ParseInt(taskIDStr, 10, 64)
			if parseErr != nil {
				http.Error(w, "Invalid task ID", http.StatusBadRequest)
				return
			}
			disputes, err = db.GetDisputesByTaskID(taskID)
		} else if status != "" {
			// Filter by status
			if status == "open" {
				disputes, err = db.GetOpenDisputes()
			} else {
				// For other statuses, we need to get all disputes and filter
				allDisputes, getErr := db.GetAllDisputes()
				if getErr != nil {
					http.Error(w, "Failed to get disputes", http.StatusInternalServerError)
					return
				}
				for _, d := range allDisputes {
					if d.Status == status {
						disputes = append(disputes, d)
					}
				}
			}
		} else {
			// Default: get all open disputes
			disputes, err = db.GetOpenDisputes()
		}

		if err != nil {
			http.Error(w, "Failed to get disputes", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"disputes": disputes,
		})
	})
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
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/disputes/assign [post]
func AssignDisputeHandler() http.HandlerFunc {
	return server.RequirePermission(server.PermDisputeManage)(func(w http.ResponseWriter, r *http.Request) {
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
	})
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
// @Failure 403 {object} map[string]interface{} "Dispute not assigned to you or insufficient permissions"
// @Failure 404 {object} map[string]interface{} "Dispute or task not found"
// @Failure 500 {object} map[string]interface{} "Failed to resolve dispute"
// @Router /api/disputes/resolve [post]
// @Security BearerAuth
func ResolveDisputeHandler() http.HandlerFunc {
	return server.RequirePermission(server.PermDisputeManage)(func(w http.ResponseWriter, r *http.Request) {
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

		// Use transaction to ensure atomicity
		tx, err := db.Postgres.Beginx()
		if err != nil {
			http.Error(w, "Failed to start transaction: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		userWallet, err := models.GetWalletByUserAndCurrency(db.Postgres, walletUserID, task.Currency)
		if err != nil {
			http.Error(w, "Failed to get user wallet: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := userWallet.AddBalance(tx, amount); err != nil {
			http.Error(w, "Failed to credit user: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := db.UpdateEscrowBalanceStatusTx(tx, task.ID, newEscrowStatus); err != nil {
			http.Error(w, "Failed to update escrow status", http.StatusInternalServerError)
			return
		}

		if err := db.UpdateDisputeStatusTx(tx, req.DisputeID, "resolved", &req.Resolution); err != nil {
			http.Error(w, "Failed to update dispute status", http.StatusInternalServerError)
			return
		}

		if err := db.UpdateTaskStatusTx(tx, task.ID, "completed"); err != nil {
			http.Error(w, "Failed to update task status", http.StatusInternalServerError)
			return
		}

		// Update freelancer's completed tasks count if freelancer won
		if req.Resolution == "freelancer_won" {
			_, err = tx.Exec(`
				UPDATE profiles
				SET completed_tasks = completed_tasks + 1
				WHERE user_id = $1
			`, walletUserID)
			if err != nil {
				http.Error(w, "Failed to update freelancer completed tasks", http.StatusInternalServerError)
				return
			}
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, "Failed to commit transaction: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Dispute resolved and funds transferred successfully",
		})
	})
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

		limit := 50
		offset := 0
		if lStr := r.URL.Query().Get("limit"); lStr != "" {
			if l, err := strconv.Atoi(lStr); err == nil && l > 0 && l <= 1000 {
				limit = l
			}
		}
		if oStr := r.URL.Query().Get("offset"); oStr != "" {
			if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
				offset = o
			}
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

		messages, err := db.GetDisputeMessagesPaged(disputeID, limit, offset)
		if err != nil {
			http.Error(w, "Failed to get dispute messages", http.StatusInternalServerError)
			return
		}

		var adminInfo interface{} = nil
		if dispute.AssignedAdmin != nil {
			admin, err := db.GetUserByID(*dispute.AssignedAdmin)
			if err == nil {
				adminTitle := ""
				if admin.AdminTitle.Valid {
					adminTitle = admin.AdminTitle.String
				}
				adminInfo = map[string]interface{}{
					"id":       admin.ID,
					"username": admin.Username,
					"title":    adminTitle,
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"dispute":  dispute,
			"task":     task,
			"escrow":   escrow,
			"messages": messages,
			"admin":    adminInfo,
		})
	}
}

// GetDisputeDetailsForUserHandler godoc
// @Summary Get dispute details for user
// @Description Returns dispute info for dispute participants (client or freelancer)
// @Tags disputes
// @Produce json
// @Param id query int true "Dispute ID"
// @Success 200 {object} map[string]interface{} "success flag and dispute details"
// @Failure 400 {string} string "Invalid dispute ID"
// @Failure 403 {string} string "Access denied - not a dispute participant"
// @Failure 404 {string} string "Dispute not found"
// @Failure 500 {string} string "Failed to get data"
// @Router /api/disputes/details [get]
// @Security BearerAuth
func GetDisputeDetailsForUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		claims := server.GetUserFromContext(r)
		if claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		userID := claims.UserID

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

		// Check if user is participant (client or freelancer)
		offers, err := db.GetTaskOffersByTaskID(db.Postgres, dispute.TaskID)
		if err != nil {
			http.Error(w, "Failed to get offers", http.StatusInternalServerError)
			return
		}

		isParticipant := task.ClientID == userID // Client is always participant
		if !isParticipant {
			// Check if user is accepted freelancer
			for _, offer := range offers {
				if offer.FreelancerID == userID && offer.Accepted {
					isParticipant = true
					break
				}
			}
		}

		if !isParticipant {
			http.Error(w, "Access denied - not a dispute participant", http.StatusForbidden)
			return
		}

		escrow, err := db.GetEscrowBalanceByTaskID(dispute.TaskID)
		if err != nil {
			http.Error(w, "Escrow balance not found", http.StatusNotFound)
			return
		}

		messages, err := db.GetDisputeMessagesPaged(disputeID, 50, 0)
		if err != nil {
			http.Error(w, "Failed to get dispute messages", http.StatusInternalServerError)
			return
		}

		var adminInfo interface{} = nil
		if dispute.AssignedAdmin != nil {
			admin, err := db.GetUserByID(*dispute.AssignedAdmin)
			if err == nil {
				adminTitle := ""
				if admin.AdminTitle.Valid {
					adminTitle = admin.AdminTitle.String
				}
				adminInfo = map[string]interface{}{
					"id":       admin.ID,
					"username": admin.Username,
					"title":    adminTitle,
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"dispute":  dispute,
			"task":     task,
			"escrow":   escrow,
			"messages": messages,
			"admin":    adminInfo,
		})
	}
}
