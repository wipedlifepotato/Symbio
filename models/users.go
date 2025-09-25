package models

import (
	"database/sql"
	"time"
)

const (
	CanChangeBalance = 1 << iota
	CanBlockUsers
	CanManageDisputes
)

type User struct {
	ID           int64          `json:"id" db:"id"`
	Username     string         `json:"username" db:"username"`
	Mnemonic     string         `json:"-" db:"mnemonic"`
	PasswordHash string         `json:"-" db:"password_hash"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
	Blocked      bool           `db:"blocked" json:"blocked"`
	IsAdmin      bool           `db:"is_admin" json:"is_admin"`
	AdminTitle   sql.NullString `db:"admin_title" json:"admin_title"`
	Permissions  int            `db:"permissions" json:"permissions"`
}
