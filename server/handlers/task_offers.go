package handlers

import (
    "encoding/json"
    "mFrelance/db"
    "mFrelance/models"
    "mFrelance/server"
    "net/http"
    "math/big"
    "strconv"
    "time"
)

type CreateTaskOfferRequest struct {
	TaskID int64   `json:"task_id"`
	Price  float64 `json:"price"`
}

type AcceptTaskOfferRequest struct {
	OfferID int64 `json:"offer_id"`
}

type CompleteTaskRequest struct {
	TaskID int64 `json:"task_id"`
}
// CreateTaskOfferHandler godoc
// @Summary Create a task offer
// @Description Allows a freelancer to make an offer on an open task
// @Tags offers
// @Accept json
// @Produce json
// @Param body body CreateTaskOfferRequest true "Offer payload"
// @Success 200 {object} map[string]interface{} "Example: {\"success\": true, \"offer\": {\"id\": 123, \"task_id\": 456, \"freelancer_id\": 78, \"price\": 50.0, \"message\": \"I can do this\", \"accepted\": false, \"created_at\": \"2023-12-01T10:00:00Z\"}}"
// @Failure 400 {string} string "Example: \"Invalid JSON\""
// @Failure 401 {string} string "Example: \"Unauthorized\""
// @Failure 404 {string} string "Example: \"Task not found\""
// @Failure 500 {string} string "Example: \"Failed to create offer\""
// @Router /api/offers [post]
// @Security BearerAuth
func CreateTaskOfferHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var offer models.TaskOffer
		if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		claims := server.GetUserFromContext(r)
		if claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
        userID := claims.UserID

        // Blocked users cannot create offers
        if blocked, err := db.IsUserBlocked(db.Postgres, userID); err != nil {
            http.Error(w, "Failed to check user status", http.StatusInternalServerError)
            return
        } else if blocked {
            http.Error(w, "User is blocked", http.StatusForbidden)
            return
        }

		task, err := db.GetTask(db.Postgres, offer.TaskID)
		if err != nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		if task.Status != "open" {
			http.Error(w, "Task is not open for offers", http.StatusBadRequest)
			return
		}

        if task.ClientID == userID {
			http.Error(w, "Cannot make offer on your own task", http.StatusBadRequest)
			return
		}

        // Limit: only one offer per user per task
        existingOffers, err := db.GetTaskOffersByTaskID(db.Postgres, offer.TaskID)
        if err != nil {
            http.Error(w, "Failed to check existing offers", http.StatusInternalServerError)
            return
        }
        for _, of := range existingOffers {
            if of.FreelancerID == userID {
                http.Error(w, "You have already made an offer for this task", http.StatusBadRequest)
                return
            }
        }

		offer.FreelancerID = userID
		offer.Accepted = false
		offer.CreatedAt = time.Now()

		if err := db.CreateTaskOffer(db.Postgres, &offer); err != nil {
			http.Error(w, "Failed to create offer", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"offer":   offer,
		})
	}
}

// UpdateTaskOfferHandler godoc
// @Summary Update user's own task offer
// @Description Allows a freelancer to update their own offer (not accepted yet)
// @Tags offers
// @Accept json
// @Produce json
// @Param offer body models.TaskOffer true "Offer data"
// @Success 200 {object} map[string]interface{} "success and updated offer"
// @Failure 400 {object} map[string]string "Invalid JSON or accepted offer cannot be edited"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden (not owner)"
// @Failure 404 {object} map[string]string "Offer not found"
// @Failure 405 {object} map[string]string "Method not allowed"
// @Failure 500 {object} map[string]string "Failed to update offer"
// @Router /api/offers/update [put]
// @Security BearerAuth
func UpdateTaskOfferHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut && r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var offer models.TaskOffer
		if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		claims := server.GetUserFromContext(r)
		if claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Fetch existing to verify ownership and status
		existing, err := db.GetTaskOffer(db.Postgres, offer.ID)
		if err != nil {
			http.Error(w, "Offer not found", http.StatusNotFound)
			return
		}
		if existing.FreelancerID != claims.UserID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		if existing.Accepted {
			http.Error(w, "Accepted offer cannot be edited", http.StatusBadRequest)
			return
		}

		// Keep immutable fields
		offer.TaskID = existing.TaskID
		offer.FreelancerID = existing.FreelancerID
		offer.Accepted = false

		if err := db.UpdateTaskOffer(db.Postgres, &offer); err != nil {
			http.Error(w, "Failed to update offer", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"offer":   offer,
		})
	}
}

// DeleteTaskOfferHandler godoc
// @Summary Delete user's own task offer
// @Description Allows a freelancer (or admin) to delete their own offer if not accepted
// @Tags offers
// @Produce json
// @Param id query int true "Offer ID"
// @Success 200 {object} map[string]interface{} "Example: {\"success\": true}"
// @Failure 400 {object} map[string]string "Example: {\"error\": \"Invalid offer ID\"}"
// @Failure 401 {object} map[string]string "Example: {\"error\": \"Unauthorized\"}"
// @Failure 403 {object} map[string]string "Example: {\"error\": \"Forbidden\"}"
// @Failure 404 {object} map[string]string "Example: {\"error\": \"Offer not found\"}"
// @Failure 405 {object} map[string]string "Example: {\"error\": \"Method not allowed\"}"
// @Failure 500 {object} map[string]string "Example: {\"error\": \"Failed to delete offer\"}"
// @Router /api/offers/delete [delete]
// @Security BearerAuth
func DeleteTaskOfferHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete && r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		offerIDStr := r.URL.Query().Get("id")
		offerID, err := strconv.ParseInt(offerIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid offer ID", http.StatusBadRequest)
			return
		}

		claims := server.GetUserFromContext(r)
		if claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		existing, err := db.GetTaskOffer(db.Postgres, offerID)
		if err != nil {
			http.Error(w, "Offer not found", http.StatusNotFound)
			return
		}
		// Owner or admin can delete
		if existing.FreelancerID != claims.UserID {
			isAdmin, err := db.IsAdmin(db.Postgres, claims.UserID)
			if err != nil || !isAdmin {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}
		if existing.Accepted {
			http.Error(w, "Accepted offer cannot be deleted", http.StatusBadRequest)
			return
		}

		if err := db.DeleteTaskOffer(db.Postgres, offerID); err != nil {
			http.Error(w, "Failed to delete offer", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	}
}
// GetTaskOffersHandler godoc
// @Summary Get task offers
// @Description Returns list of offers for a task. Freelancers can only see their own offers
// @Tags offers
// @Produce json
// @Param task_id query int true "Task ID"
// @Success 200 {object} map[string]interface{} "success flag and offers list"
// @Failure 400 {string} string "Invalid task ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 404 {string} string "Task not found"
// @Failure 500 {string} string "Failed to get offers"
// @Router /api/offers [get]
// @Security BearerAuth

func GetTaskOffersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		taskIDStr := r.URL.Query().Get("task_id")
		taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid task ID", http.StatusBadRequest)
			return
		}

		task, err := db.GetTask(db.Postgres, taskID)
		if err != nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		claims := server.GetUserFromContext(r)
		if claims == nil {
			http.Error(w, "Unauthorized1", http.StatusUnauthorized)
			return
		}
		userID := claims.UserID

		if task.ClientID != userID {

			offers, err := db.GetTaskOffersByFreelancerID(db.Postgres, userID)
			if err != nil {
				http.Error(w, "Failed to get offers", http.StatusInternalServerError)
				return
			}

			hasOffer := false
			for _, offer := range offers {
				if offer.TaskID == taskID {
					hasOffer = true
					break
				}
			}

			if !hasOffer {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}

		offers, err := db.GetTaskOffersByTaskID(db.Postgres, taskID)
		if err != nil {
			http.Error(w, "Failed to get offers", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"offers":  offers,
		})
	}
}
// AcceptTaskOfferHandler godoc
// @Summary Accept a task offer
// @Description Allows the task owner to accept a freelancer's offer, debit wallet, and put funds in escrow
// @Tags offers
// @Accept json
// @Produce json
// @Param body body AcceptTaskOfferRequest true "Offer acceptance payload"
// @Success 200 {object} map[string]interface{} "success flag and message"
// @Failure 400 {string} string "Invalid JSON, insufficient balance or bad request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 404 {string} string "Offer or task not found"
// @Failure 500 {string} string "Failed to accept offer"
// @Router /api/offers/accept [post]
// @Security BearerAuth

func AcceptTaskOfferHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			OfferID int64 `json:"offer_id"`
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
		userID := claims.UserID

		offer, err := db.GetTaskOffer(db.Postgres, req.OfferID)
		if err != nil {
			http.Error(w, "Offer not found", http.StatusNotFound)
			return
		}

		task, err := db.GetTask(db.Postgres, offer.TaskID)
		if err != nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		if task.ClientID != userID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if task.Status != "open" {
			http.Error(w, "Task is not open", http.StatusBadRequest)
			return
		}

		userWallet, err := models.GetWalletByUserAndCurrency(db.Postgres, userID, task.Currency)
		if err != nil {
			http.Error(w, "Failed to get wallet: "+err.Error(), http.StatusInternalServerError)
			return
		}

		amountBig := big.NewFloat(offer.Price)
		if userWallet.BigBalance().Cmp(amountBig) < 0 {
			http.Error(w, "Insufficient balance", http.StatusBadRequest)
			return
		}

		if err := userWallet.SubBalance(db.Postgres, amountBig); err != nil {
			http.Error(w, "Failed to debit wallet: "+err.Error(), http.StatusInternalServerError)
			return
		}

		escrow := &models.EscrowBalance{
			TaskID:       task.ID,
			ClientID:     task.ClientID,
			FreelancerID: offer.FreelancerID,
			Amount:       offer.Price,
			Currency:     task.Currency,
			Status:       "pending",
			CreatedAt:    time.Now(),
		}

		if err := db.CreateEscrowBalance(escrow); err != nil {
			http.Error(w, "Failed to create escrow: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := db.AcceptTaskOffer(db.Postgres, req.OfferID); err != nil {
			http.Error(w, "Failed to accept offer: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := db.RejectOtherOffersForTask(db.Postgres, task.ID, req.OfferID); err != nil {
			http.Error(w, "Failed to reject other offers: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := db.UpdateTaskStatus(db.Postgres, task.ID, "in_progress"); err != nil {
			http.Error(w, "Failed to update task status: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Offer accepted successfully",
		})
	}
}
// CompleteTaskHandler godoc
// @Summary Complete a task
// @Description Allows the client to confirm completion of a task and release escrow to the freelancer
// @Tags tasks
// @Accept json
// @Produce json
// @Param body body CompleteTaskRequest true "Task completion payload"
// @Success 200 {object} map[string]interface{} "success flag and message"
// @Failure 400 {string} string "Invalid JSON or bad request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 404 {string} string "Task or escrow not found"
// @Failure 500 {string} string "Failed to complete task"
// @Router /api/tasks/complete [post]
// @Security BearerAuth

func CompleteTaskHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			TaskID int64 `json:"task_id"`
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
		userID := claims.UserID

		task, err := db.GetTask(db.Postgres, req.TaskID)
		if err != nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		if task.ClientID != userID {
			http.Error(w, "Forbidden: only client can confirm completion", http.StatusForbidden)
			return
		}

		offers, err := db.GetTaskOffersByTaskID(db.Postgres, req.TaskID)
		if err != nil {
			http.Error(w, "Failed to get offers", http.StatusInternalServerError)
			return
		}

		var acceptedOffer *models.TaskOffer
		for _, offer := range offers {
			if offer.Accepted {
				acceptedOffer = offer
				break
			}
		}

		if acceptedOffer == nil {
			http.Error(w, "No accepted offer for this task", http.StatusBadRequest)
			return
		}

		if task.Status != "in_progress" {
			http.Error(w, "Task is not in progress", http.StatusBadRequest)
			return
		}

		escrow, err := db.GetEscrowBalanceByTaskID(task.ID)
		if err != nil {
			http.Error(w, "Escrow balance not found", http.StatusNotFound)
			return
		}

		amountBig := big.NewFloat(escrow.Amount)

		freelancerWallet, err := models.GetWalletByUserAndCurrency(db.Postgres, acceptedOffer.FreelancerID, task.Currency)
		if err != nil {
			http.Error(w, "Failed to get freelancer wallet: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Use transaction to ensure atomicity
		tx, err := db.Postgres.Beginx()
		if err != nil {
			http.Error(w, "Failed to start transaction: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		if err := freelancerWallet.AddBalance(tx, amountBig); err != nil {
			http.Error(w, "Failed to credit freelancer: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := db.UpdateEscrowBalanceStatusTx(tx, task.ID, "released"); err != nil {
			http.Error(w, "Failed to update escrow status", http.StatusInternalServerError)
			return
		}

		if err := db.UpdateTaskStatusTx(tx, task.ID, "completed"); err != nil {
			http.Error(w, "Failed to update task status", http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec(`
		    UPDATE profiles
		    SET completed_tasks = completed_tasks + 1
		    WHERE user_id = $1
		`, acceptedOffer.FreelancerID)
		if err != nil {
		    http.Error(w, "Failed to update freelancer completed tasks", http.StatusInternalServerError)
		    return
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, "Failed to commit transaction: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Task confirmed by client and completed successfully",
		})
	}
}
