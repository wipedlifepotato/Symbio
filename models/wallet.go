package models

import (
	"fmt"
	"log"
	"math/big"

	"github.com/jmoiron/sqlx"
)

type Wallet struct {
	ID       int64  `db:"id"`
	UserID   int64  `db:"user_id"`
	Balance  string `db:"balance"`
	Currency string `db:"currency"`
	Address  string `db:"address"`
}

func GetWalletsByUser(db *sqlx.DB, userID int64) ([]Wallet, error) {
	rows, err := db.Query(`
		SELECT id, user_id, balance, currency, address
		FROM wallets
		WHERE user_id=$1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wallets []Wallet
	for rows.Next() {
		var w Wallet
		if err := rows.Scan(&w.ID, &w.UserID, &w.Balance, &w.Currency, &w.Address); err != nil {
			continue
		}
		wallets = append(wallets, w)
	}
	return wallets, nil
}

func GetWalletByUserAndCurrency(db *sqlx.DB, userID int64, currency string) (*Wallet, error) {
	var w Wallet
	err := db.QueryRow(`
		SELECT id, user_id, balance, currency, address
		FROM wallets
		WHERE user_id=$1 AND currency=$2
		LIMIT 1
	`, userID, currency).Scan(&w.ID, &w.UserID, &w.Balance, &w.Currency, &w.Address)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func GetWalletByAddress(db *sqlx.DB, address, currency string) (*Wallet, error) {
	var w Wallet
	err := db.QueryRow(`
		SELECT id, user_id, balance, currency, address
		FROM wallets
		WHERE address=$1 AND currency=$2
		LIMIT 1
	`, address, currency).Scan(&w.ID, &w.UserID, &w.Balance, &w.Currency, &w.Address)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func IsOurWalletAddress(db *sqlx.DB, address, currency string) (bool, error) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM wallets WHERE address=$1 AND currency=$2)
	`, address, currency).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (w *Wallet) BigBalance() *big.Float {
	b, ok := new(big.Float).SetString(w.Balance)
	if !ok {
		log.Printf("invalid balance format for wallet %d: %s", w.ID, w.Balance)
		return big.NewFloat(0)
	}
	return b
}

func (w *Wallet) SetBalance(db interface{}, newBal *big.Float) error {
	newBalanceStr := fmt.Sprintf("%.8f", newBal)
	var err error
	if tx, ok := db.(*sqlx.Tx); ok {
		_, err = tx.Exec(`
			UPDATE wallets
			SET balance=$1
			WHERE id=$2
		`, newBalanceStr, w.ID)
	} else if dbConn, ok := db.(*sqlx.DB); ok {
		_, err = dbConn.Exec(`
			UPDATE wallets
			SET balance=$1
			WHERE id=$2
		`, newBalanceStr, w.ID)
	} else {
		return fmt.Errorf("unsupported database interface")
	}
	if err == nil {
		w.Balance = newBalanceStr
	}
	return err
}

func (w *Wallet) AddBalance(db interface{}, delta *big.Float) error {
	current := w.BigBalance()
	newBal := new(big.Float).Add(current, delta)
	return w.SetBalance(db, newBal)
}

func (w *Wallet) SubBalance(db interface{}, delta *big.Float) error {
	current := w.BigBalance()
	newBal := new(big.Float).Sub(current, delta)
	if newBal.Cmp(big.NewFloat(0)) < 0 {
		return fmt.Errorf("insufficient balance")
	}
	return w.SetBalance(db, newBal)
}

func AddToWalletBalance(db *sqlx.DB, address, currency string, delta *big.Float) error {
	w, err := GetWalletByAddress(db, address, currency)
	if err != nil {
		return err
	}
	return w.AddBalance(db, delta)
}
