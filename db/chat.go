package db

import (
	"mFrelance/models"

	"github.com/jmoiron/sqlx"
)

// CreateChatRoom creates a new chat room
func CreateChatRoom(db *sqlx.DB) (*models.ChatRoom, error) {
	room := &models.ChatRoom{}
	err := db.Get(room, `INSERT INTO chat_rooms DEFAULT VALUES RETURNING id, created_at`)
	if err != nil {
		return nil, err
	}
	return room, nil
}

// GetChatRoom retrieves a chat room by ID
func GetChatRoom(db *sqlx.DB, id int64) (*models.ChatRoom, error) {
	var room models.ChatRoom
	err := db.Get(&room, `SELECT * FROM chat_rooms WHERE id = $1`, id)
	return &room, err
}

// CreateChatParticipant adds a user to a chat room
func CreateChatParticipant(db *sqlx.DB, participant *models.ChatParticipant) error {
	_, err := db.NamedExec(`INSERT INTO chat_participants (chat_room_id, user_id, joined_at) VALUES (:chat_room_id, :user_id, :joined_at)`, participant)
	return err
}

// GetChatParticipants retrieves participants in a chat room
func GetChatParticipants(db *sqlx.DB, chatRoomID int64) ([]models.ChatParticipant, error) {
	var participants []models.ChatParticipant
	err := db.Select(&participants, `SELECT * FROM chat_participants WHERE chat_room_id = $1`, chatRoomID)
	return participants, err
}

// CreateChatMessage adds a message to a chat room
func CreateChatMessage(db *sqlx.DB, message *models.ChatMessage) error {
	_, err := db.NamedExec(`INSERT INTO chat_messages (chat_room_id, sender_id, message, created_at) VALUES (:chat_room_id, :sender_id, :message, :created_at)`, message)
	return err
}

// GetChatMessages retrieves messages in a chat room
func GetChatMessages(db *sqlx.DB, chatRoomID int64) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	err := db.Select(&messages, `SELECT * FROM chat_messages WHERE chat_room_id = $1 ORDER BY created_at`, chatRoomID)
	return messages, err
}

// CreateChatRequest creates a new chat request
func CreateChatRequest(db *sqlx.DB, request *models.ChatRequest) error {
	_, err := db.NamedExec(`INSERT INTO chat_requests (requester_id, requested_id, status, created_at) VALUES (:requester_id, :requested_id, :status, :created_at)`, request)
	return err
}

// GetChatRequest retrieves a chat request
func GetChatRequest(db *sqlx.DB, requesterID, requestedID int64) (*models.ChatRequest, error) {
	var request models.ChatRequest
	err := db.Get(&request, `SELECT * FROM chat_requests WHERE requester_id = $1 AND requested_id = $2`, requesterID, requestedID)
	return &request, err
}

// UpdateChatRequest updates the status of a chat request
func UpdateChatRequest(db *sqlx.DB, request *models.ChatRequest) error {
	_, err := db.NamedExec(`UPDATE chat_requests SET status = :status WHERE requester_id = :requester_id AND requested_id = :requested_id`, request)
	return err
}

// GetChatRoomsForUser retrieves chat rooms for a given user
func GetChatRoomsForUser(db *sqlx.DB, userID int64) ([]models.ChatRoom, error) {
	var chatRooms []models.ChatRoom
	err := db.Select(&chatRooms, `
        SELECT cr.* FROM chat_rooms cr
        INNER JOIN chat_participants cp ON cr.id = cp.chat_room_id
        WHERE cp.user_id = $1
    `, userID)
	return chatRooms, err
}

func IsUserHaveAccessToChatRoom(db *sqlx.DB, userID int64, chatRoomID int64) (bool, error) {
	var count int
	err := db.Get(&count, `
		SELECT COUNT(*) FROM chat_rooms cr
		INNER JOIN chat_participants cp ON cr.id = cp.chat_room_id
		WHERE cp.user_id = $1 AND cr.id = $2
	`, userID, chatRoomID)
	return count > 0, err
}

func GetUsersInChatRoom(db *sqlx.DB, chatRoomID int64) ([]models.User, error) {
	var users []models.User
	err := db.Select(&users, `
		SELECT u.* FROM users u
		INNER JOIN chat_participants cp ON u.id = cp.user_id
		WHERE cp.chat_room_id = $1
	`, chatRoomID)
	return users, err
}

func AddUserToChatRoom(db *sqlx.DB, userID int64, chatRoomID int64) error {
	_, err := db.Exec(`
		INSERT INTO chat_participants (chat_room_id, user_id, joined_at)
		VALUES ($1, $2, NOW())
	`, chatRoomID, userID)
	return err
}
