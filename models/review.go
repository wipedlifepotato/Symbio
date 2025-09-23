package models

import (
	"time"
)

type Review struct {
	ID         int64     `db:"id" json:"id"`
	TaskID     int64     `db:"task_id" json:"task_id"`
	ReviewerID int64     `db:"reviewer_id" json:"reviewer_id"`
	ReviewedID int64     `db:"reviewed_id" json:"reviewed_id"`
	Rating     int       `db:"rating" json:"rating"`
	Comment    string    `db:"comment" json:"comment"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type ReviewResponse struct {
	ID        int64     `db:"id" json:"id"`
	ReviewID  int64     `db:"review_id" json:"review_id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	Response  string    `db:"response" json:"response"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

