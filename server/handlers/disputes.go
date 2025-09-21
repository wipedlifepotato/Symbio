package handlers

import (
	"encoding/json"
	"mFrelance/db"
	"mFrelance/models"
	"mFrelance/server"
	"net/http"
	"strconv"
	"time"
)

func CreateDisputeHandler() http.HandlerFunc {
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
			http.Error(w, "Unauthorized1", http.StatusUnauthorized)
			return
		}
		userID := claims.UserID

		task, err := db.GetTask(db.Postgres, req.TaskID)
		if err != nil {
			http.Error(w, "Task not found", http.StatusNotFound)
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
			http.Error(w, "No accepted offer found", http.StatusBadRequest)
			return
		}

		if task.ClientID != userID && acceptedOffer.FreelancerID != userID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if task.Status != "in_progress" {
			http.Error(w, "Task is not in progress", http.StatusBadRequest)
			return
		}

		existingDisputes, err := db.GetDisputesByTaskID(req.TaskID)
		if err != nil {
			http.Error(w, "Failed to check existing disputes", http.StatusInternalServerError)
			return
		}

		if len(existingDisputes) > 0 {
			http.Error(w, "Dispute already exists for this task", http.StatusBadRequest)
			return
		}

		dispute := &models.Dispute{
			TaskID:    req.TaskID,
			OpenedBy:  userID,
			Status:    "open",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := db.CreateDispute(dispute); err != nil {
			http.Error(w, "Failed to create dispute", http.StatusInternalServerError)
			return
		}

		if err := db.UpdateTaskStatus(db.Postgres, task.ID, "disputed"); err != nil {
			http.Error(w, "Failed to update task status", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"dispute": dispute,
		})
	}
}

func GetDisputeHandler() http.HandlerFunc {
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

		messages, err := db.GetDisputeMessages(disputeID)
		if err != nil {
			http.Error(w, "Failed to get dispute messages", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"dispute":  dispute,
			"messages": messages,
		})
	}
}

func SendDisputeMessageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			DisputeID int64  `json:"dispute_id"`
			Message   string `json:"message"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		claims := server.GetUserFromContext(r)
		if claims == nil {
			http.Error(w, "Unauthorized1", http.StatusUnauthorized)
			return
		}
		userID := claims.UserID

		dispute, err := db.GetDisputeByID(req.DisputeID)
		if err != nil {
			http.Error(w, "Dispute not found", http.StatusNotFound)
			return
		}

		task, err := db.GetTask(db.Postgres, dispute.TaskID)
		if err != nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		offers, err := db.GetTaskOffersByTaskID(db.Postgres, dispute.TaskID)
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
			http.Error(w, "No accepted offer found", http.StatusBadRequest)
			return
		}

		if task.ClientID != userID && acceptedOffer.FreelancerID != userID && dispute.AssignedAdmin == nil || (dispute.AssignedAdmin != nil && *dispute.AssignedAdmin != userID) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		message := &models.DisputeMessage{
			DisputeID: req.DisputeID,
			SenderID:  userID,
			Message:   req.Message,
			CreatedAt: time.Now(),
		}

		if err := db.CreateDisputeMessage(message); err != nil {
			http.Error(w, "Failed to create message", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": message,
		})
	}
}

func GetUserDisputesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		claims := server.GetUserFromContext(r)
		if claims == nil {
			http.Error(w, "Unauthorized1", http.StatusUnauthorized)
			return
		}
		userID := claims.UserID

		// Получаем все задачи пользователя
		userTasks, err := db.GetTasksByClientID(db.Postgres, userID)
		if err != nil {
			http.Error(w, "Failed to get user tasks", http.StatusInternalServerError)
			return
		}

		var disputes []*models.Dispute
		for _, task := range userTasks {
			taskDisputes, err := db.GetDisputesByTaskID(task.ID)
			if err != nil {
				continue
			}
			disputes = append(disputes, taskDisputes...)
		}

		offers, err := db.GetTaskOffersByFreelancerID(db.Postgres, userID)
		if err != nil {
			http.Error(w, "Failed to get user offers", http.StatusInternalServerError)
			return
		}

		for _, offer := range offers {
			if offer.Accepted {
				taskDisputes, err := db.GetDisputesByTaskID(offer.TaskID)
				if err != nil {
					continue
				}
				disputes = append(disputes, taskDisputes...)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"disputes": disputes,
		})
	}
}
