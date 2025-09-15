package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type Transaction struct {
	ID           int64          `db:"id"`
	FromWalletID sql.NullInt64  `db:"from_wallet_id"`
	ToWalletID   sql.NullInt64  `db:"to_wallet_id"`
	ToAddress    sql.NullString `db:"to_address"`
	TaskID       sql.NullInt64  `db:"task_id"`
	Amount       string         `db:"amount"`
	Currency     string         `db:"currency"`
	Confirmed    bool           `db:"confirmed"`
	CreatedAt    time.Time      `db:"created_at"`
}

func SaveTransaction(db *sqlx.DB, tx *Transaction) error {
	_, err := db.Exec(`
        INSERT INTO transactions (from_wallet_id, to_wallet_id, to_address, task_id, amount, currency, confirmed, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `,
		tx.FromWalletID, tx.ToWalletID, tx.ToAddress, tx.TaskID, tx.Amount, tx.Currency, tx.Confirmed, time.Now(),
	)
	return err
}

func GetTransaction(db *sqlx.DB, id int64) (*Transaction, error) {
	tx := &Transaction{}
	err := db.QueryRow(`
        SELECT id, from_wallet_id, to_wallet_id, to_address, task_id, amount, currency, confirmed, created_at
        FROM transactions
        WHERE id = $1
        LIMIT 1
    `, id).Scan(
		&tx.ID,
		&tx.FromWalletID,
		&tx.ToWalletID,
		&tx.ToAddress,
		&tx.TaskID,
		&tx.Amount,
		&tx.Currency,
		&tx.Confirmed,
		&tx.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // транзакция не найдена
		}
		return nil, fmt.Errorf("GetTransaction error: %w", err)
	}
	return tx, nil
}

func GetTransactions(db *sqlx.DB, limit int, offset int) ([]*Transaction, error) {
	rows, err := db.Query(`
        SELECT id, from_wallet_id, to_wallet_id, to_address, task_id, amount, currency, confirmed, created_at
        FROM transactions
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("GetTransactions error: %w", err)
	}
	defer rows.Close()

	var txs []*Transaction
	for rows.Next() {
		tx := &Transaction{}
		if err := rows.Scan(
			&tx.ID,
			&tx.FromWalletID,
			&tx.ToWalletID,
			&tx.ToAddress,
			&tx.TaskID,
			&tx.Amount,
			&tx.Currency,
			&tx.Confirmed,
			&tx.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("GetTransactions scan error: %w", err)
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

func GetTransactionsByWallet(db *sqlx.DB, walletID int64, limit, offset int) ([]*Transaction, error) {
	rows, err := db.Query(`
        SELECT id, from_wallet_id, to_wallet_id, to_address, task_id, amount, currency, confirmed, created_at
        FROM transactions
        WHERE from_wallet_id=$1 OR to_wallet_id=$1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `, walletID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []*Transaction
	for rows.Next() {
		tx := new(Transaction)
		if err := rows.Scan(
			&tx.ID, &tx.FromWalletID, &tx.ToWalletID, &tx.ToAddress, &tx.TaskID, &tx.Amount,
			&tx.Currency, &tx.Confirmed, &tx.CreatedAt,
		); err != nil {
			continue
		}
		txs = append(txs, tx)
	}
	return txs, nil
}
