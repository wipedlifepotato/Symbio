package db

import (
	"mFrelance/models"

	"github.com/jmoiron/sqlx"
	//"mFrelance/models"
)

func CreateMessage(db *sqlx.DB, message *models.Message) error {
	_, err := db.NamedExec(`INSERT INTO messages (sender_id, recipient_id, task_id, message, read, created_at) VALUES (:sender_id, :recipient_id, :task_id, :message, :read, :created_at)`, message)
	return err
}

func GetMessage(db *sqlx.DB, id int64) (*models.Message, error) {
	var message models.Message
	err := db.Get(&message, `SELECT * FROM messages WHERE id = $1`, id)
	return &message, err
}

func UpdateMessage(db *sqlx.DB, message *models.Message) error {
	_, err := db.NamedExec(`UPDATE messages SET sender_id=:sender_id, recipient_id=:recipient_id, task_id=:task_id, message=:message, read=:read WHERE id=:id`, message)
	return err
}

func DeleteMessage(db *sqlx.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM messages WHERE id = $1`, id)
	return err
}
