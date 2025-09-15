package models

import "time"

type User struct {
	ID           int64     `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Mnemonic     string    `json:"-" db:"mnemonic"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	Blocked      bool      `db:"blocked" json:"blocked"`
	IsAdmin      bool      `db:"is_admin" json:"is_admin"`
}
