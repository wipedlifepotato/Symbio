package db

import (
	"database/sql"
	"time"
	"errors"
	"github.com/jmoiron/sqlx"
)


func CheckUser(db *sqlx.DB, username, passwordHash string) (int64, error) {
	var userID int64
	err := db.QueryRow(`
        SELECT id FROM users
        WHERE username = $1 AND password_hash = $2
    `, username, passwordHash).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // пользователь не найден
		}
		return 0, err
	}
	return userID, nil
}

//   err = db.RestoreUser(db.Postgres, req.Username, req.Mnemonic)
func RestoreUser(db *sqlx.DB, wantusername, mnemonic string) (int64, string, error) {
	var (
		userID   int64
		username string
	)
	err := db.QueryRow(`
        SELECT id, username FROM users
        WHERE mnemonic = $1
    `, mnemonic).Scan(&userID, &username)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, "", nil
		}
		return 0, "", err
	}

	if wantusername != username {
		return 0, "", errors.New("username does not match mnemonic")
	}
	return userID, username, nil
}

func ChangeUserPassword(db *sqlx.DB, username, passwordHash string) error {
	res, err := db.Exec(`
        UPDATE users
        SET password_hash = $1
        WHERE username = $2
    `, passwordHash, username)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("user not found")
	}

	return nil
}

func CreateUser(db *sqlx.DB, username, passwordHash, mnemonic string) error {
    _, err := db.Exec(`
        INSERT INTO users (username, password_hash, mnemonic, created_at)
        VALUES ($1, $2, $3, $4)
    `, username, passwordHash, mnemonic, time.Now())
    return err
}

func GetUserByUsername(db *sqlx.DB, username string) (int64, string, error) {
	var (
		userID       int64
		passwordHash string
	)
	err := db.QueryRow(`
        SELECT id, password_hash FROM users WHERE username = $1
    `, username).Scan(&userID, &passwordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, "", nil
		}
		return 0, "", err
	}
	return userID, passwordHash, nil
}

