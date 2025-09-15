package db

import (
	"github.com/jmoiron/sqlx"
	"mFrelance/models"
)

func CreateTaskOffer(db *sqlx.DB, offer *models.TaskOffer) error {
	_, err := db.NamedExec(`INSERT INTO task_offers (task_id, freelancer_id, price, message, status, created_at) VALUES (:task_id, :freelancer_id, :price, :message, :status, :created_at)`, offer)
	return err
}

func GetTaskOffer(db *sqlx.DB, id int64) (*models.TaskOffer, error) {
	var offer models.TaskOffer
	err := db.Get(&offer, `SELECT * FROM task_offers WHERE id = $1`, id)
	return &offer, err
}

func UpdateTaskOffer(db *sqlx.DB, offer *models.TaskOffer) error {
	_, err := db.NamedExec(`UPDATE task_offers SET task_id=:task_id, freelancer_id=:freelancer_id, price=:price, message=:message, status=:status WHERE id=:id`, offer)
	return err
}

func DeleteTaskOffer(db *sqlx.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM task_offers WHERE id = $1`, id)
	return err
}
