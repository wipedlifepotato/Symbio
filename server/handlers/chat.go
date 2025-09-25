package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"database/sql"
	"mFrelance/db"
	"mFrelance/models"
	"mFrelance/server"
	time "time"
)

// CreateChatRequestHandler creates a new chat request
// @Summary Create a chat request
// @Description Create a new chat request from the logged-in user to another user. 
// If a request already exists and is pending or open, a new request cannot be created.
// @Tags Chat
// @Accept json
// @Produce json
// @Param requested_id query int true "ID of the user you want to start a chat with"
// @Success 201 {object} models.ChatRequest "Returns the created chat request"
// @Failure 400 {string} string "Invalid requested_id or request already exists and is open/pending"
// @Failure 401 {string} string "Unauthorized — user not logged in"
// @Failure 500 {string} string "Database error"
// @Router /chat/createChatRequest [post]
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

		// Проверяем, есть ли уже открытый запрос между пользователями
		existingRequest, err := db.GetChatRequest(db.Postgres, requestedID, claims.UserID)
		if err != nil && err != sql.ErrNoRows {
			http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if existingRequest != nil && (existingRequest.Status == "pending" || existingRequest.Status == "open") {
			http.Error(w, "request already exists and is open/pending", http.StatusBadRequest)
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
// @Summary Update a chat request
// @Description Accept or reject a chat request by the requested user
// @Tags Chat
// @Accept json
// @Produce json
// @Param requester_id query int true "ID of the user who sent the chat request"
// @Param status query string true "New status: accepted or rejected"
// @Success 200 {object} map[string]string "Returns status ok"
// @Failure 400 {string} string "Invalid requester_id or status"
// @Failure 401 {string} string "Unauthorized — user not logged in or not allowed to update"
// @Failure 500 {string} string "Database error"
// @Router /chat/UpdateChatRequest [post]
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

// GetChatRoomsForUserHandler retrieves chat rooms for the logged-in user
// @Summary Get chat rooms
// @Description Returns all chat rooms the logged-in user participates in with participant info
// @Tags Chat
// @Accept json
// @Produce json
// @Success 200 {array} object "List of chat rooms with participant usernames"
// @Failure 401 {string} string "Unauthorized — user not logged in"
// @Failure 500 {string} string "Database error"
// @Router /chat/getChatRoomsForUser [get]
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
		log.Print("chatRooms:")
		log.Println(chatRooms)
		// Get participant info for each room
		type ChatRoomInfo struct {
			ID        int64     `json:"id"`
			Name      string    `json:"name"`      // For display: "Chat username"
			Username  string    `json:"username"`  // The other participant's username
			UserID    int64     `json:"user_id"`   // The other participant's ID
			CreatedAt time.Time `json:"created_at"`
		}

		result := make([]ChatRoomInfo, 0)
		for _, room := range chatRooms {
			users, err := db.GetUsersInChatRoom(db.Postgres, room.ID)
			log.Printf("Room %d has %d users", room.ID, len(users))
			if err != nil {
				log.Printf("Error getting users for room %d: %v", room.ID, err)
				continue
			}

			// Find the other participant
			foundOther := false
			for _, user := range users {
				log.Printf("User in room %d: ID=%d, Username=%s, CurrentUserID=%d", room.ID, user.ID, user.Username, claims.UserID)
				if user.ID != claims.UserID {
					result = append(result, ChatRoomInfo{
						ID:        room.ID,
						Name:      "Chat " + user.Username,
						Username:  user.Username,
						UserID:    user.ID,
						CreatedAt: room.CreatedAt,
					})
					log.Printf("Added chat room %d with user %s", room.ID, user.Username)
					foundOther = true
					break
				}
			}

			// If no other participant found, show as "Empty Chat" or similar
			if !foundOther && len(users) > 0 {
				// At least current user is in the room
				result = append(result, ChatRoomInfo{
					ID:        room.ID,
					Name:      "Empty Chat",
					Username:  "",
					UserID:    0,
					CreatedAt: room.CreatedAt,
				})
				log.Printf("Added empty chat room %d", room.ID)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

// GetChatMessagesHandler retrieves messages for a chat room
// @Summary Get chat messages
// @Description Returns all messages for a given chat room if the user has access
// @Tags Chat
// @Accept json
// @Produce json
// @Param chat_room_id query int true "ID of the chat room"
// @Success 200 {array} models.ChatMessage "List of messages"
// @Failure 400 {string} string "Invalid chat_room_id"
// @Failure 401 {string} string "Unauthorized — user not logged in or no access"
// @Failure 500 {string} string "Database error"
// @Router /chat/getChatMessages [get]
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

		if h, err := db.IsUserHaveAccessToChatRoom(db.Postgres, claims.UserID, chatRoomID); err != nil {
			http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
			return
		} else if h == false {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		messages, err := db.GetChatMessagesPaged(db.Postgres, chatRoomID, limit, offset)
		if err != nil {
			http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(messages)
	}
}

// SendMessageHandler sends a message to a chat room
// @Summary Send message
// @Description Sends a message to a chat room for the logged-in user
// @Tags Chat
// @Accept json
// @Produce json
// @Param chat_room_id query int true "ID of the chat room"
// @Param message body models.ChatMessage true "Message object"
// @Success 201 {object} models.ChatMessage "Returns the created message"
// @Failure 400 {string} string "Invalid chat_room_id or request body"
// @Failure 401 {string} string "Unauthorized — user not logged in"
// @Failure 500 {string} string "Database error"
// @Router /chat/sendMessage [post]
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

// GetChatRequestsHandler retrieves chat requests for the logged-in user
// @Summary Get chat requests
// @Description Returns all incoming chat requests for the logged-in user
// @Tags Chat
// @Accept json
// @Produce json
// @Success 200 {array} models.ChatRequest "List of chat requests"
// @Failure 401 {string} string "Unauthorized — user not logged in"
// @Failure 500 {string} string "Database error"
// @Router /chat/getChatRequests [get]
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

// AcceptChatRequestHandler accepts a chat request and creates a chat room
// @Summary Accept chat request
// @Description Accepts a chat request from another user and creates a chat room
// @Tags Chat
// @Accept json
// @Produce json
// @Param requester_id query int true "ID of the user who sent the request"
// @Success 200 {object} map[string]interface{} "Returns accepted status and chatRoomID"
// @Failure 400 {string} string "Invalid requester_id"
// @Failure 401 {string} string "Unauthorized — user not logged in"
// @Failure 500 {string} string "Database error"
// @Router /chat/acceptChatRequest [post]
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

// ExitFromChat allows a user to leave a chat room
// @Summary Exit chat room
// @Description Removes the logged-in user from a chat room
// @Tags Chat
// @Accept json
// @Produce json
// @Param chat_room_id query int true "ID of the chat room"
// @Success 200 {object} map[string]string "Returns status ok"
// @Failure 400 {string} string "Invalid chat_room_id"
// @Failure 401 {string} string "Unauthorized — user not logged in"
// @Failure 500 {string} string "Database error"
// @Router /chat/exitFromChat [post]
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

// CancelChatRequestHandler cancels a chat request sent by the logged-in user
// @Summary Cancel chat request
// @Description Cancels a previously sent chat request
// @Tags Chat
// @Accept json
// @Produce json
// @Param requester_id query int true "ID of the user to whom the request was sent"
// @Success 200 {string} string "Request cancelled successfully"
// @Failure 400 {string} string "Invalid requester_id"
// @Failure 401 {string} string "Unauthorized — user not logged in"
// @Failure 500 {string} string "Database error"
// @Router /chat/cancelChatRequest [post]
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
