package models

import (
	"time"
)

type Task struct {
	ID          int64     `db:"id" json:"id"`
	ClientID    int64     `db:"client_id" json:"client_id"`
	Title       string    `db:"title" json:"title"`
	Description string    `db:"description" json:"description"`
	Category    string    `db:"category" json:"category"`
	Budget      float64   `db:"budget" json:"budget"`
	Currency    string    `db:"currency" json:"currency"`
	Status      string    `db:"status" json:"status"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	Deadline    time.Time `db:"deadline" json:"deadline"`
}
