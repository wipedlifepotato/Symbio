package db

import (
	"github.com/jmoiron/sqlx"
	"mFrelance/models"
)

func CreateTask(db *sqlx.DB, task *models.Task) error {
	_, err := db.NamedExec(`INSERT INTO tasks (client_id, title, description, category, budget, currency, status, created_at, deadline)
	VALUES (:client_id, :title, :description, :category, :budget, :currency, :status, :created_at, :deadline)`, task)
	return err
}

func GetTask(db *sqlx.DB, id int64) (*models.Task, error) {
	var task models.Task
	err := db.Get(&task, `SELECT * FROM tasks WHERE id = $1`, id)
	return &task, err
}

func UpdateTask(db *sqlx.DB, task *models.Task) error {
	_, err := db.NamedExec(`UPDATE tasks SET client_id=:client_id, title=:title, description=:description, category=:category, budget=:budget, currency=:currency, status=:status, deadline=:deadline WHERE id=:id`, task)
	return err
}

func DeleteTask(db *sqlx.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM tasks WHERE id = $1`, id)
	return err
}
