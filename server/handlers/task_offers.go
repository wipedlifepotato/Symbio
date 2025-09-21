package handlers

import (
	"encoding/json"
	"mFrelance/db"
	"mFrelance/models"
	"mFrelance/server"
	"math/big"
	"net/http"
	"strconv"
	"time"
)

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

		if err := freelancerWallet.AddBalance(db.Postgres, amountBig); err != nil {
			http.Error(w, "Failed to credit freelancer: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := db.UpdateEscrowBalanceStatus(task.ID, "released"); err != nil {
			http.Error(w, "Failed to update escrow status", http.StatusInternalServerError)
			return
		}

		if err := db.UpdateTaskStatus(db.Postgres, task.ID, "completed"); err != nil {
			http.Error(w, "Failed to update task status", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Task confirmed by client and completed successfully",
		})
	}
}
