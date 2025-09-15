package models

import (
	"time"
)

type TaskOffer struct {
	ID           int64     `db:"id" json:"id"`
	TaskID       int64     `db:"task_id" json:"task_id"`
	FreelancerID int64     `db:"freelancer_id" json:"freelancer_id"`
	Price        float64   `db:"price" json:"price"`
	Message      string    `db:"message" json:"message"`
	Status       string    `db:"status" json:"status"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
