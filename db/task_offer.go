package db

import (
	"github.com/jmoiron/sqlx"
	"mFrelance/models"
)

func CreateTaskOffer(db *sqlx.DB, offer *models.TaskOffer) error {
	_, err := db.NamedExec(`INSERT INTO task_offers (task_id, freelancer_id, price, message, accepted, created_at) VALUES (:task_id, :freelancer_id, :price, :message, :accepted, :created_at)`, offer)
	return err
}

func GetTaskOffer(db *sqlx.DB, id int64) (*models.TaskOffer, error) {
	var offer models.TaskOffer
	err := db.Get(&offer, `SELECT * FROM task_offers WHERE id = $1`, id)
	return &offer, err
}

func GetTaskOffersByTaskID(db *sqlx.DB, taskID int64) ([]*models.TaskOffer, error) {
	var offers []*models.TaskOffer
	err := db.Select(&offers, `SELECT * FROM task_offers WHERE task_id = $1 ORDER BY created_at DESC`, taskID)
	return offers, err
}

func GetTaskOffersByFreelancerID(db *sqlx.DB, freelancerID int64) ([]*models.TaskOffer, error) {
	var offers []*models.TaskOffer
	err := db.Select(&offers, `SELECT * FROM task_offers WHERE freelancer_id = $1 ORDER BY created_at DESC`, freelancerID)
	return offers, err
}

func AcceptTaskOffer(db *sqlx.DB, offerID int64) error {
	_, err := db.Exec(`UPDATE task_offers SET accepted = true WHERE id = $1`, offerID)
	return err
}

func RejectOtherOffersForTask(db *sqlx.DB, taskID, acceptedOfferID int64) error {
	_, err := db.Exec(`UPDATE task_offers SET accepted = false WHERE task_id = $1 AND id != $2`, taskID, acceptedOfferID)
	return err
}

func UpdateTaskOffer(db *sqlx.DB, offer *models.TaskOffer) error {
	_, err := db.NamedExec(`UPDATE task_offers SET task_id=:task_id, freelancer_id=:freelancer_id, price=:price, message=:message, accepted=:accepted WHERE id=:id`, offer)
	return err
}

func DeleteTaskOffer(db *sqlx.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM task_offers WHERE id = $1`, id)
	return err
}
