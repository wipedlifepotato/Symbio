package models

import (
	"time"
)

type Message struct {
	ID          int64     `db:"id" json:"id"`
	SenderID    int64     `db:"sender_id" json:"sender_id"`
	RecipientID int64     `db:"recipient_id" json:"recipient_id"`
	TaskID      int64     `db:"task_id" json:"task_id"`
	Message     string    `db:"message" json:"message"`
	Read        bool      `db:"read" json:"read"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}
