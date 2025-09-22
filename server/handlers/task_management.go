package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"mFrelance/db"
	"mFrelance/models"
	"mFrelance/server"
)

type CreateTaskRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
	Deadline    string  `json:"deadline"` // ISO8601
}

type UpdateTaskRequest struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
	Deadline    string  `json:"deadline"`
}

// CreateTaskHandler godoc
// @Summary Create a new task
// @Description Allows a user to create a new task
// @Tags tasks
// @Accept json
// @Produce json
// @Param body body CreateTaskRequest true "Task payload"
// @Success 200 {object} map[string]interface{} "success flag and created task"
// @Failure 400 {string} string "Invalid JSON"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Failed to create task"
// @Router /api/tasks [post]
// @Security BearerAuth
func CreateTaskHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[CreateTaskHandler] Starting task creation request")

		if r.Method != http.MethodPost {
			log.Printf("[CreateTaskHandler] Invalid method: %s", r.Method)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var task models.Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			log.Printf("[CreateTaskHandler] JSON decode error: %v", err)
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		log.Printf("[CreateTaskHandler] Decoded task: %+v", task)

		claims := server.GetUserFromContext(r)
		if claims == nil {
			log.Printf("[CreateTaskHandler] No user claims in context")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID := claims.UserID

		log.Printf("[CreateTaskHandler] User ID: %d", userID)

		task.ClientID = userID
		task.Status = "open"
		task.CreatedAt = time.Now()

		log.Printf("[CreateTaskHandler] Task before creation: %+v", task)

		if err := db.CreateTask(db.Postgres, &task); err != nil {
			log.Printf("[CreateTaskHandler] Database error: %v", err)
			http.Error(w, "Failed to create task", http.StatusInternalServerError)
			return
		}

		log.Printf("[CreateTaskHandler] Task created successfully with ID: %d", task.ID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"task":    task,
		})
	}
}
// GetTasksHandler godoc
// @Summary Get tasks
// @Description Returns list of tasks. Use query param `status=open` to get only open tasks
// @Tags tasks
// @Produce json
// @Param status query string false "Filter tasks by status"
// @Success 200 {object} map[string]interface{} "success flag and tasks list"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Failed to get tasks"
// @Router /api/tasks [get]
// @Security BearerAuth
func GetTasksHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		//log.Println("get tasks")
		status := r.URL.Query().Get("status")
		var tasks []*models.Task
		var err error

		if status == "open" {
			//	log.Println("Get opet tasks")
			tasks, err = db.GetOpenTasks(db.Postgres)
			if err != nil {
				log.Printf("GetOpenTasks error: %v", err)
			} else {
				log.Printf("GetOpenTasks count=%d", len(tasks))
				for i, t := range tasks {
					log.Printf("[%d] %+v", i, t)
				}
			}
			log.Println(tasks)
		} else {

			claims := server.GetUserFromContext(r)
			if claims == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			userID := claims.UserID
			tasks, err = db.GetTasksByClientID(db.Postgres, userID)
		}

		if err != nil {
			http.Error(w, "Failed to get tasks", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"tasks":   tasks,
		})
	}
}
// GetTaskHandler godoc
// @Summary Get task details
// @Description Returns details of a single task
// @Tags tasks
// @Produce json
// @Param id query int true "Task ID"
// @Success 200 {object} map[string]interface{} "success flag and task"
// @Failure 400 {string} string "Invalid task ID"
// @Failure 404 {string} string "Task not found"
// @Router /api/tasks/detail [get]
// @Security BearerAuth
func GetTaskHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		taskIDStr := r.URL.Query().Get("id")
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

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"task":    task,
		})
	}
}
// UpdateTaskHandler godoc
// @Summary Update a task
// @Description Allows the task owner to update a task
// @Tags tasks
// @Accept json
// @Produce json
// @Param body body UpdateTaskRequest true "Updated task payload"
// @Success 200 {object} map[string]interface{} "success flag and updated task"
// @Failure 400 {string} string "Invalid JSON"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 404 {string} string "Task not found"
// @Failure 500 {string} string "Failed to update task"
// @Router /api/tasks [put]
// @Security BearerAuth
func UpdateTaskHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var task models.Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		claims := server.GetUserFromContext(r)
		if claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		userID := claims.UserID

		existingTask, err := db.GetTask(db.Postgres, task.ID)
		if err != nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		if existingTask.ClientID != userID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if err := db.UpdateTask(db.Postgres, &task); err != nil {
			http.Error(w, "Failed to update task", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"task":    task,
		})
	}
}
// DeleteTaskHandler godoc
// @Summary Delete a task
// @Description Allows the task owner to delete a task
// @Tags tasks
// @Produce json
// @Param id query int true "Task ID"
// @Success 200 {object} map[string]interface{} "success flag"
// @Failure 400 {string} string "Invalid task ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 404 {string} string "Task not found"
// @Failure 500 {string} string "Failed to delete task"
// @Router /api/tasks [delete]
// @Security BearerAuth
func DeleteTaskHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		taskIDStr := r.URL.Query().Get("id")
		taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid task ID", http.StatusBadRequest)
			return
		}

		claims := server.GetUserFromContext(r)
		if claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		userID := claims.UserID

		existingTask, err := db.GetTask(db.Postgres, taskID)
		if err != nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		if existingTask.ClientID != userID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if err := db.DeleteTask(db.Postgres, taskID); err != nil {
			http.Error(w, "Failed to delete task", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	}
}
