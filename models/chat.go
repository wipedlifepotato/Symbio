package models

import "time"

type ChatRoom struct {
	ID        int64     `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type ChatParticipant struct {
	ID         int64     `db:"id" json:"id"`
	ChatRoomID int64     `db:"chat_room_id" json:"chat_room_id"`
	UserID     int64     `db:"user_id" json:"user_id"`
	JoinedAt   time.Time `db:"joined_at" json:"joined_at"`
}

type ChatMessage struct {
	ID         int64     `db:"id" json:"id"`
	ChatRoomID int64     `db:"chat_room_id" json:"chat_room_id"`
	SenderID   int64     `db:"sender_id" json:"sender_id"`
	Message    string    `db:"message" json:"message"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type ChatRequest struct {
	ID          int64     `db:"id" json:"id"`
	RequesterID int64     `db:"requester_id" json:"requester_id"`
	RequestedID int64     `db:"requested_id" json:"requested_id"`
	Status      string    `db:"status" json:"status"` // pending, accepted, rejected
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}
