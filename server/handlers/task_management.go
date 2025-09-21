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
