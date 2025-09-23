package models

import (
	"time"
)

type Dispute struct {
	ID            int64     `db:"id" json:"id"`
	TaskID        int64     `db:"task_id" json:"task_id"`
	OpenedBy      int64     `db:"opened_by" json:"opened_by"`
	AssignedAdmin *int64    `db:"assigned_admin" json:"assigned_admin"`
	Status        string    `db:"status" json:"status"`
	Resolution    *string   `db:"resolution" json:"resolution"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

type DisputeMessage struct {
	ID        int64     `db:"id" json:"id"`
	DisputeID int64     `db:"dispute_id" json:"dispute_id"`
	SenderID  int64     `db:"sender_id" json:"sender_id"`
	Message   string    `db:"message" json:"message"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

