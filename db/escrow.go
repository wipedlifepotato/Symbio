package db

import (
	"mFrelance/models"
	// "time"
)

func CreateEscrowBalance(escrow *models.EscrowBalance) error {
	query := `
		INSERT INTO escrow_balances (task_id, client_id, freelancer_id, amount, currency, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	return Postgres.QueryRow(query, escrow.TaskID, escrow.ClientID, escrow.FreelancerID, escrow.Amount, escrow.Currency, escrow.Status, escrow.CreatedAt).Scan(&escrow.ID)
}

func GetEscrowBalanceByTaskID(taskID int64) (*models.EscrowBalance, error) {
	escrow := &models.EscrowBalance{}
	query := `SELECT id, task_id, client_id, freelancer_id, amount, currency, status, created_at FROM escrow_balances WHERE task_id = $1`

	err := Postgres.QueryRow(query, taskID).Scan(
		&escrow.ID, &escrow.TaskID, &escrow.ClientID, &escrow.FreelancerID,
		&escrow.Amount, &escrow.Currency, &escrow.Status, &escrow.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return escrow, nil
}

func UpdateEscrowBalanceStatus(taskID int64, status string) error {
	query := `UPDATE escrow_balances SET status = $1 WHERE task_id = $2`
	_, err := Postgres.Exec(query, status, taskID)
	return err
}

func GetEscrowBalancesByUserID(userID int64) ([]*models.EscrowBalance, error) {
	query := `SELECT id, task_id, client_id, freelancer_id, amount, currency, status, created_at FROM escrow_balances WHERE client_id = $1 OR freelancer_id = $1 ORDER BY created_at DESC`
	rows, err := Postgres.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var balances []*models.EscrowBalance
	for rows.Next() {
		balance := &models.EscrowBalance{}
		err := rows.Scan(
			&balance.ID, &balance.TaskID, &balance.ClientID, &balance.FreelancerID,
			&balance.Amount, &balance.Currency, &balance.Status, &balance.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		balances = append(balances, balance)
	}
	return balances, nil
}

