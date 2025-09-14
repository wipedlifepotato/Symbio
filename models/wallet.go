package models

import (
	//"database/sql"
	"fmt"
	"log"
	"math/big"
	"mFrelance/db"
)

type Wallet struct {
	ID       int64  `db:"id"`
	UserID   int64  `db:"user_id"`
	Balance  string `db:"balance"`
	Currency string `db:"currency"`
	Address  string `db:"address"`
}

func GetWalletsByUser(userID int64) ([]Wallet, error) {
	rows, err := db.Postgres.Query(`
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

func GetWalletByUserAndCurrency(userID int64, currency string) (*Wallet, error) {
	var w Wallet
	err := db.Postgres.QueryRow(`
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

func GetWalletByAddress(address, currency string) (*Wallet, error) {
	var w Wallet
	err := db.Postgres.QueryRow(`
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

func IsOurWalletAddress(address, currency string) (bool, error) {
	var exists bool
	err := db.Postgres.QueryRow(`
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

func (w *Wallet) SetBalance(newBal *big.Float) error {
	newBalanceStr := fmt.Sprintf("%.8f", newBal)
	_, err := db.Postgres.Exec(`
		UPDATE wallets 
		SET balance=$1 
		WHERE id=$2
	`, newBalanceStr, w.ID)
	if err == nil {
		w.Balance = newBalanceStr
	}
	return err
}

func (w *Wallet) AddBalance(delta *big.Float) error {
	current := w.BigBalance()
	newBal := new(big.Float).Add(current, delta)
	return w.SetBalance(newBal)
}

func (w *Wallet) SubBalance(delta *big.Float) error {
	current := w.BigBalance()
	newBal := new(big.Float).Sub(current, delta)
	if newBal.Cmp(big.NewFloat(0)) < 0 {
		return fmt.Errorf("insufficient balance")
	}
	return w.SetBalance(newBal)
}

func AddToWalletBalance(address, currency string, delta *big.Float) error {
	w, err := GetWalletByAddress(address, currency)
	if err != nil {
		return err
	}
	return w.AddBalance(delta)
}

