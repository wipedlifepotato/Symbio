package models

import (
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
