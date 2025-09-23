package db

import (
	"log"
	"mFrelance/models"

	"github.com/jmoiron/sqlx"
)

func GetEscrowByTaskID(tx *sqlx.Tx, taskID int64) (*models.EscrowBalance, error) {
	var escrow models.EscrowBalance
	err := tx.Get(&escrow, `SELECT * FROM escrow_balances WHERE task_id=$1`, taskID)
	if err != nil {
		return nil, err
	}
	return &escrow, nil
}

func UpdateEscrowBalanceStatusTx(tx *sqlx.Tx, taskID int64, status string) error {
	_, err := tx.Exec(`UPDATE escrow_balances SET status=$1 WHERE task_id=$2`, status, taskID)
	return err
}

func UpdateTaskStatusTx(tx *sqlx.Tx, taskID int64, status string) error {
	_, err := tx.Exec(`UPDATE tasks SET status=$1 WHERE id=$2`, status, taskID)
	return err
}

func CreateTask(db *sqlx.DB, task *models.Task) error {
	log.Printf("[CreateTask] Creating task: %+v", task)

	deadlineTime := task.Deadline.Time

	query := `
		INSERT INTO tasks (
			client_id, title, description, category, budget, currency, status, created_at, deadline
		) VALUES (
			:client_id, :title, :description, :category, :budget, :currency, :status, :created_at, :deadline
		)
		RETURNING id
	`

	params := map[string]interface{}{
		"client_id":   task.ClientID,
		"title":       task.Title,
		"description": task.Description,
		"category":    task.Category,
		"budget":      task.Budget,
		"currency":    task.Currency,
		"status":      task.Status,
		"created_at":  task.CreatedAt,
		"deadline":    deadlineTime,
	}

	stmt, err := db.PrepareNamed(query)
	if err != nil {
		log.Printf("[CreateTask] PrepareNamed error: %v", err)
		return err
	}
	defer stmt.Close()

	err = stmt.Get(&task.ID, params)
	if err != nil {
		log.Printf("[CreateTask] Database error: %v", err)
		return err
	}

	log.Printf("[CreateTask] Task created successfully with ID=%d", task.ID)
	return nil
}

func GetTask(db *sqlx.DB, id int64) (*models.Task, error) {
	var task models.Task
	err := db.Get(&task, `SELECT * FROM tasks WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func GetTasksByClientID(db *sqlx.DB, clientID int64) ([]*models.Task, error) {
	var tasks []*models.Task
	err := db.Select(&tasks, `SELECT * FROM tasks WHERE client_id = $1 ORDER BY created_at DESC`, clientID)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func GetOpenTasks(db *sqlx.DB) ([]*models.Task, error) {
	var tasks []*models.Task
	err := db.Select(&tasks, `SELECT * FROM tasks WHERE status = 'open' ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func CountOpenTasks(db *sqlx.DB) (int64, error) {
    var n int64
    err := db.Get(&n, `SELECT COUNT(*) FROM tasks WHERE status='open'`)
    return n, err
}

func CountTasksByClientAndStatus(db *sqlx.DB, clientID int64, status string) (int64, error) {
    var n int64
    err := db.Get(&n, `SELECT COUNT(*) FROM tasks WHERE client_id=$1 AND status=$2`, clientID, status)
    return n, err
}

// Paged fetching helpers
func GetTasksByClientIDPaged(db *sqlx.DB, clientID int64, limit, offset int) ([]*models.Task, error) {
    var tasks []*models.Task
    err := db.Select(&tasks, `SELECT * FROM tasks WHERE client_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, clientID, limit, offset)
    if err != nil {
        return nil, err
    }
    return tasks, nil
}

func GetOpenTasksPaged(db *sqlx.DB, limit, offset int) ([]*models.Task, error) {
    var tasks []*models.Task
    err := db.Select(&tasks, `SELECT * FROM tasks WHERE status = 'open' ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
    if err != nil {
        return nil, err
    }
    return tasks, nil
}

func GetTasksByClientIDAndStatusPaged(db *sqlx.DB, clientID int64, status string, limit, offset int) ([]*models.Task, error) {
    var tasks []*models.Task
    query := `SELECT * FROM tasks WHERE client_id = $1 AND status = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`
    err := db.Select(&tasks, query, clientID, status, limit, offset)
    if err != nil {
        return nil, err
    }
    return tasks, nil
}

func UpdateTask(db *sqlx.DB, task *models.Task) error {
	query := `
		UPDATE tasks
		SET client_id = :client_id,
			title = :title,
			description = :description,
			category = :category,
			budget = :budget,
			currency = :currency,
			status = :status,
			deadline = :deadline
		WHERE id = :id
	`
	_, err := db.NamedExec(query, task)
	return err
}

func UpdateTaskStatus(db *sqlx.DB, taskID int64, status string) error {
	_, err := db.Exec(`UPDATE tasks SET status = $1 WHERE id = $2`, status, taskID)
	return err
}

func DeleteTask(db *sqlx.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM tasks WHERE id = $1`, id)
	return err
}

func DeleteTasksByUserID(db *sqlx.DB, userID int64) (int64, error) {
    res, err := db.Exec(`DELETE FROM tasks WHERE client_id = $1`, userID)
    if err != nil {
        return 0, err
    }
    n, _ := res.RowsAffected()
    return n, nil
}
