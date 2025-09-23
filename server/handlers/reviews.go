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
// CreateReviewHandler godoc
// @Summary Create a review
// @Description Create a review for a completed task. Only the client or the accepted freelancer can review.
// @Tags reviews
// @Accept json
// @Produce json
// @Param review body models.Review true "Review payload"
// @Success 200 {object} models.Review
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Task not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /api/reviews [post]
func CreateReviewHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var review models.Review
		if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		claims := server.GetUserFromContext(r)
		if claims == nil {
			http.Error(w, "Unauthorized1", http.StatusUnauthorized)
			return
		}
		userID := claims.UserID

		task, err := db.GetTask(db.Postgres, review.TaskID)
		if err != nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		if task.Status != "completed" {
			http.Error(w, "Task is not completed", http.StatusBadRequest)
			return
		}

		offers, err := db.GetTaskOffersByTaskID(db.Postgres, review.TaskID)
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

		hasReviewed, err := db.HasUserReviewedTask(userID, review.TaskID)
		if err != nil {
			http.Error(w, "Failed to check existing reviews", http.StatusInternalServerError)
			return
		}

		if hasReviewed {
			http.Error(w, "Review already exists for this task", http.StatusBadRequest)
			return
		}

		review.ReviewerID = userID
		review.CreatedAt = time.Now()

		if task.ClientID == userID {
			review.ReviewedID = acceptedOffer.FreelancerID
		} else {
			review.ReviewedID = task.ClientID
		}

		if review.Rating < 1 || review.Rating > 5 {
			http.Error(w, "Rating must be between 1 and 5", http.StatusBadRequest)
			return
		}

		if err := db.CreateReview(&review); err != nil {
			http.Error(w, "Failed to create review", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"review":  review,
		})
	}
}
// CreateReviewHandler godoc
// @Summary Create a review
// @Description Create a review for a completed task. Only the client or the accepted freelancer can review.
// @Tags reviews
// @Accept json
// @Produce json
// @Param review body models.Review true "Review payload"
// @Success 200 {object} models.Review
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Task not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /api/reviews [post]
func GetReviewsByUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userIDStr := r.URL.Query().Get("user_id")
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		reviews, err := db.GetReviewsByUserID(userID)
		if err != nil {
			http.Error(w, "Failed to get reviews", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"reviews": reviews,
		})
	}
}
// GetReviewsByTaskHandler godoc
// @Summary Get reviews by task
// @Description Returns all reviews for a specific task
// @Tags reviews
// @Produce json
// @Param task_id query int true "Task ID"
// @Success 200 {array} models.Review
// @Failure 400 {object} map[string]string "Invalid task ID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/reviews/by-task [get]
func GetReviewsByTaskHandler() http.HandlerFunc {
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

		reviews, err := db.GetReviewsByTaskID(taskID)
		if err != nil {
			http.Error(w, "Failed to get reviews", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"reviews": reviews,
		})
	}
}
// GetUserRatingHandler godoc
// @Summary Get user rating
// @Description Returns the average rating of a user
// @Tags reviews
// @Produce json
// @Param user_id query int true "User ID"
// @Success 200 {object} map[string]float64 "Rating"
// @Failure 400 {object} map[string]string "Invalid user ID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/reviews/rating [get]
func GetUserRatingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userIDStr := r.URL.Query().Get("user_id")
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		rating, err := db.GetUserRating(userID)
		if err != nil {
			http.Error(w, "Failed to get rating", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"rating":  rating,
		})
	}
}

