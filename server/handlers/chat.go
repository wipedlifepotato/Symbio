package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"mFrelance/db"
	"mFrelance/models"
	"mFrelance/server"
	time "time"
)

// CreateChatRequestHandler creates a new chat request
func CreateChatRequestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := server.GetUserFromContext(r)
		if claims == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		requestedIDStr := r.URL.Query().Get("requested_id")
		if requestedIDStr == "" {
			http.Error(w, "invalid requested_id", http.StatusBadRequest)
			return
		}

		requestedID, err := strconv.ParseInt(requestedIDStr, 10, 64)
		if err != nil || requestedID <= 0 {
			http.Error(w, "invalid requested_id", http.StatusBadRequest)
			return
		}

		request := &models.ChatRequest{
			RequesterID: requestedID,
			RequestedID: claims.UserID,
			Status:      "pending",
			CreatedAt:   time.Now(),
		}

		err = db.CreateChatRequest(db.Postgres, request)
		if err != nil {
			http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(request)
	}
}

// UpdateChatRequestHandler updates a chat request (accept/reject)
func UpdateChatRequestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := server.GetUserFromContext(r)
		if claims == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		requesterIDStr := r.URL.Query().Get("requester_id")
		if requesterIDStr == "" {
			http.Error(w, "invalid requester_id", http.StatusBadRequest)
			return
		}

		requesterID, err := strconv.ParseInt(requesterIDStr, 10, 64)
		if err != nil || requesterID <= 0 {
			http.Error(w, "invalid requester_id", http.StatusBadRequest)
			return
		}

		newStatus := r.URL.Query().Get("status")
		if newStatus != "accepted" && newStatus != "rejected" {
			http.Error(w, "invalid status", http.StatusBadRequest)
			return
		}

		// Only the requested user can accept/reject
		if claims.UserID != requesterID {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		request := &models.ChatRequest{
			RequesterID: requesterID,
			RequestedID: claims.UserID,
			Status:      newStatus,
		}

		err = db.UpdateChatRequest(db.Postgres, request)
		if err != nil {
			http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// If accepted, create chat room and add participants
		if newStatus == "accepted" {
			room, err := db.CreateChatRoom(db.Postgres)
			if err != nil {
				http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
				return
			}

			participant1 := &models.ChatParticipant{
				ChatRoomID: room.ID,
				UserID:     requesterID,
				JoinedAt:   time.Now(),
			}
			err = db.CreateChatParticipant(db.Postgres, participant1)
			if err != nil {
				http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
				return
			}

			participant2 := &models.ChatParticipant{
				ChatRoomID: room.ID,
				UserID:     claims.UserID,
				JoinedAt:   time.Now(),
			}
			err = db.CreateChatParticipant(db.Postgres, participant2)
			if err != nil {
				http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

// GetChatRoomsForUserHandler retrieves chat rooms for a given user
func GetChatRoomsForUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := server.GetUserFromContext(r)
		if claims == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		chatRooms, err := db.GetChatRoomsForUser(db.Postgres, claims.UserID)
		if err != nil {
			http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(chatRooms)
	}
}

// GetChatMessagesHandler retrieves messages for a chat room
func GetChatMessagesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := server.GetUserFromContext(r)
		if claims == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		chatRoomIDStr := r.URL.Query().Get("chat_room_id")
		if chatRoomIDStr == "" {
			http.Error(w, "invalid chat_room_id", http.StatusBadRequest)
			return
		}

		chatRoomID, err := strconv.ParseInt(chatRoomIDStr, 10, 64)
		if err != nil || chatRoomID <= 0 {
			http.Error(w, "invalid chat_room_id", http.StatusBadRequest)
			return
		}

		if h, err := db.IsUserHaveAccessToChatRoom(db.Postgres, claims.UserID, chatRoomID); err != nil {
			http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
			return
		} else if h == false {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		messages, err := db.GetChatMessages(db.Postgres, chatRoomID)
		if err != nil {
			http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(messages)
	}
}

// SendMessageHandler sends a message to a chat room
func SendMessageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := server.GetUserFromContext(r)
		if claims == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		chatRoomIDStr := r.URL.Query().Get("chat_room_id")
		if chatRoomIDStr == "" {
			http.Error(w, "invalid chat_room_id", http.StatusBadRequest)
			return
		}

		chatRoomID, err := strconv.ParseInt(chatRoomIDStr, 10, 64)
		if err != nil || chatRoomID <= 0 {
			http.Error(w, "invalid chat_room_id", http.StatusBadRequest)
			return
		}

		var message models.ChatMessage
		err = json.NewDecoder(r.Body).Decode(&message)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		message.ChatRoomID = chatRoomID
		message.SenderID = claims.UserID
		message.CreatedAt = time.Now()

		err = db.CreateChatMessage(db.Postgres, &message)
		if err != nil {
			http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(message)
	}
}

func GetChatRequestsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := server.GetUserFromContext(r)
		if claims == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		requests, err := db.GetChatRequestsForUser(db.Postgres, claims.UserID)
		if err != nil {
			http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(requests)
	}
}

func AcceptChatRequestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := server.GetUserFromContext(r)
		if claims == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		requesterIDStr := r.URL.Query().Get("requester_id")
		if requesterIDStr == "" {
			http.Error(w, "invalid requester_id", http.StatusBadRequest)
			return
		}

		requesterID, err := strconv.ParseInt(requesterIDStr, 10, 64)
		if err != nil || requesterID <= 0 {
			http.Error(w, "invalid requester_id", http.StatusBadRequest)
			return
		}
		if requesterID == claims.UserID {
			writeErrorJSON(w, "invalid requester_id", http.StatusBadRequest)
			return
		}
		err = db.AcceptChatRequest(db.Postgres, requesterID, claims.UserID)
		if err != nil {
			http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		db.DeleteChatRequest(db.Postgres, claims.UserID, requesterID)
		room, err := db.CreateChatRoom(db.Postgres)
		if err != nil {
			http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		participants := []struct {
			UserID int64
		}{
			{UserID: requesterID},
			{UserID: claims.UserID},
		}

		for _, p := range participants {
			cp := &models.ChatParticipant{
				ChatRoomID: room.ID,
				UserID:     p.UserID,
				JoinedAt:   time.Now(),
			}
			if err := db.CreateChatParticipant(db.Postgres, cp); err != nil {
				http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":     "accepted",
			"chatRoomID": room.ID,
		})
	}
}

func ExitFromChat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("ExitFromChat handler called")

		claims := server.GetUserFromContext(r)
		if claims == nil {
			log.Println("unauthorized: claims nil")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		chatRoomID := r.URL.Query().Get("chat_room_id")
		log.Printf("chatRoomID: %s, UserID: %d", chatRoomID, claims.UserID)

		val, err := strconv.ParseInt(chatRoomID, 10, 64)
		if err != nil {
			log.Println("invalid chat_room_id")
			http.Error(w, "invalid chat_room_id", http.StatusBadRequest)
			return
		}

		err = db.DeleteChatParticipant(db.Postgres, val, claims.UserID)
		if err != nil {
			log.Printf("db error: %v", err)
			http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("Deleted participant chat_room_id=%d user_id=%d", val, claims.UserID)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
		})
	}
}

func CancelChatRequestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := server.GetUserFromContext(r)
		if claims == nil {
			log.Println("unauthorized: claims nil")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		requesterIDStr := r.URL.Query().Get("requester_id")
		if requesterIDStr == "" {
			http.Error(w, "invalid requester_id", http.StatusBadRequest)
			return
		}
		requesterID, err := strconv.ParseInt(requesterIDStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid requester_id", http.StatusBadRequest)
			return
		}
		db.DeleteChatRequest(db.Postgres, claims.UserID, requesterID)
	}
}
