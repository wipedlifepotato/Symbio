package db

import (
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
		
)

type WalletBalance struct {
	Address string  `json:"address"`
	Balance float64 `json:"balance"`
}

func GetWalletBalance(db *sqlx.DB, userID int64, currency string) (*WalletBalance, error) {
	var wallet WalletBalance
	err := db.QueryRow(`SELECT address, balance FROM wallets WHERE user_id=$1 AND currency=$2`, userID, currency).
		Scan(&wallet.Address, &wallet.Balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &wallet, nil
}

func IsOurAddr(db *sqlx.DB, addr string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM wallets WHERE address=$1)`
	err := db.Get(&exists, query, addr)
	if err != nil {
		return false, err
	}
	return exists, nil
}
