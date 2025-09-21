package models

import (
	"time"
)

type EscrowBalance struct {
	ID           int64     `db:"id" json:"id"`
	TaskID       int64     `db:"task_id" json:"task_id"`
	ClientID     int64     `db:"client_id" json:"client_id"`
	FreelancerID int64     `db:"freelancer_id" json:"freelancer_id"`
	Amount       float64   `db:"amount" json:"amount"`
	Currency     string    `db:"currency" json:"currency"`
	Status       string    `db:"status" json:"status"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
