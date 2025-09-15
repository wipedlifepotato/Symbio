package db

import (
	"github.com/jmoiron/sqlx"
	"mFrelance/models"
)

func SaveTransaction(db *sqlx.DB, tx *models.Transaction) error {
	return models.SaveTransaction(db, tx)
}

func GetTransaction(db *sqlx.DB, id int64) (*models.Transaction, error) {
	return models.GetTransaction(db, id)
}

func GetTransactions(db *sqlx.DB, limit int, offset int) ([]*models.Transaction, error) {
	return models.GetTransactions(db, limit, offset)
}

func GetTransactionsByWallet(db *sqlx.DB, walletID int64, limit, offset int) ([]*models.Transaction, error) {
	return models.GetTransactionsByWallet(db, walletID, limit, offset)
}
