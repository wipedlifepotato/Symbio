package db

import (
	"database/sql"
	"mFrelance/models"
	"time"
)

func CreateDispute(dispute *models.Dispute) error {
	query := `
		INSERT INTO disputes (task_id, opened_by, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	return Postgres.QueryRow(query, dispute.TaskID, dispute.OpenedBy, dispute.Status, dispute.CreatedAt, dispute.UpdatedAt).Scan(&dispute.ID)
}

func GetDisputeByID(id int64) (*models.Dispute, error) {
	dispute := &models.Dispute{}
	query := `SELECT id, task_id, opened_by, assigned_admin, status, resolution, created_at, updated_at FROM disputes WHERE id = $1`

	err := Postgres.QueryRow(query, id).Scan(
		&dispute.ID, &dispute.TaskID, &dispute.OpenedBy, &dispute.AssignedAdmin,
		&dispute.Status, &dispute.Resolution, &dispute.CreatedAt, &dispute.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return dispute, nil
}

func GetDisputesByTaskID(taskID int64) ([]*models.Dispute, error) {
	query := `SELECT id, task_id, opened_by, assigned_admin, status, resolution, created_at, updated_at FROM disputes WHERE task_id = $1 ORDER BY created_at DESC`
	rows, err := Postgres.Query(query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var disputes []*models.Dispute
	for rows.Next() {
		dispute := &models.Dispute{}
		err := rows.Scan(
			&dispute.ID, &dispute.TaskID, &dispute.OpenedBy, &dispute.AssignedAdmin,
			&dispute.Status, &dispute.Resolution, &dispute.CreatedAt, &dispute.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		disputes = append(disputes, dispute)
	}
	return disputes, nil
}

func UpdateDisputeStatus(id int64, status string, resolution *string) error {
	query := `UPDATE disputes SET status = $1, resolution = $2, updated_at = $3 WHERE id = $4`
	_, err := Postgres.Exec(query, status, resolution, time.Now(), id)
	return err
}

func AssignDisputeToAdmin(disputeID, adminID int64) error {
	query := `UPDATE disputes SET assigned_admin = $1, updated_at = $2 WHERE id = $3`
	_, err := Postgres.Exec(query, adminID, time.Now(), disputeID)
	return err
}

func CreateDisputeMessage(message *models.DisputeMessage) error {
	query := `
		INSERT INTO dispute_messages (dispute_id, sender_id, message, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	return Postgres.QueryRow(query, message.DisputeID, message.SenderID, message.Message, message.CreatedAt).Scan(&message.ID)
}

func GetDisputeMessages(disputeID int64) ([]*models.DisputeMessage, error) {
	return GetDisputeMessagesPaged(disputeID, 0, 0)
}

func GetDisputeMessagesPaged(disputeID int64, limit, offset int) ([]*models.DisputeMessage, error) {
	query := `SELECT id, dispute_id, sender_id, message, created_at FROM dispute_messages WHERE dispute_id = $1 ORDER BY created_at DESC`
	var rows *sql.Rows
	var err error
	if limit > 0 {
		query += ` LIMIT $2 OFFSET $3`
		rows, err = Postgres.Query(query, disputeID, limit, offset)
	} else {
		rows, err = Postgres.Query(query, disputeID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.DisputeMessage
	for rows.Next() {
		message := &models.DisputeMessage{}
		err := rows.Scan(&message.ID, &message.DisputeID, &message.SenderID, &message.Message, &message.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, nil
}

func GetOpenDisputes() ([]*models.Dispute, error) {
	query := `SELECT id, task_id, opened_by, assigned_admin, status, resolution, created_at, updated_at FROM disputes WHERE status = 'open' ORDER BY created_at DESC`
	rows, err := Postgres.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var disputes []*models.Dispute
	for rows.Next() {
		dispute := &models.Dispute{}
		err := rows.Scan(
			&dispute.ID, &dispute.TaskID, &dispute.OpenedBy, &dispute.AssignedAdmin,
			&dispute.Status, &dispute.Resolution, &dispute.CreatedAt, &dispute.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		disputes = append(disputes, dispute)
	}
	return disputes, nil
}

func GetAllDisputes() ([]*models.Dispute, error) {
	query := `SELECT id, task_id, opened_by, assigned_admin, status, resolution, created_at, updated_at FROM disputes ORDER BY created_at DESC`
	rows, err := Postgres.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var disputes []*models.Dispute
	for rows.Next() {
		dispute := &models.Dispute{}
		err := rows.Scan(
			&dispute.ID, &dispute.TaskID, &dispute.OpenedBy, &dispute.AssignedAdmin,
			&dispute.Status, &dispute.Resolution, &dispute.CreatedAt, &dispute.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		disputes = append(disputes, dispute)
	}
	return disputes, nil
}

