package db

import (
	"database/sql"
	"mFrelance/models"
	// "time"
)

func CreateReview(review *models.Review) error {
	query := `
		INSERT INTO reviews (task_id, reviewer_id, reviewed_id, rating, comment, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	return Postgres.QueryRow(query, review.TaskID, review.ReviewerID, review.ReviewedID, review.Rating, review.Comment, review.CreatedAt).Scan(&review.ID)
}

func GetReviewsByUserID(userID int64) ([]*models.Review, error) {
	query := `SELECT id, task_id, reviewer_id, reviewed_id, rating, comment, created_at FROM reviews WHERE reviewed_id = $1 ORDER BY created_at DESC`
	rows, err := Postgres.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []*models.Review
	for rows.Next() {
		review := &models.Review{}
		err := rows.Scan(&review.ID, &review.TaskID, &review.ReviewerID, &review.ReviewedID, &review.Rating, &review.Comment, &review.CreatedAt)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	return reviews, nil
}

func GetReviewsByTaskID(taskID int64) ([]*models.Review, error) {
	query := `SELECT id, task_id, reviewer_id, reviewed_id, rating, comment, created_at FROM reviews WHERE task_id = $1 ORDER BY created_at DESC`
	rows, err := Postgres.Query(query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []*models.Review
	for rows.Next() {
		review := &models.Review{}
		err := rows.Scan(&review.ID, &review.TaskID, &review.ReviewerID, &review.ReviewedID, &review.Rating, &review.Comment, &review.CreatedAt)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	return reviews, nil
}

func GetUserRating(userID int64) (float64, error) {
	query := `SELECT AVG(rating) FROM reviews WHERE reviewed_id = $1`
	var rating sql.NullFloat64
	err := Postgres.QueryRow(query, userID).Scan(&rating)
	if err != nil {
		return 0, err
	}
	if rating.Valid {
		return rating.Float64, nil
	}
	return 0, nil
}

func HasUserReviewedTask(userID, taskID int64) (bool, error) {
	query := `SELECT COUNT(*) FROM reviews WHERE reviewer_id = $1 AND task_id = $2`
	var count int
	err := Postgres.QueryRow(query, userID, taskID).Scan(&count)
	return count > 0, err
}
