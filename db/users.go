package db

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func BlockUser(db *sqlx.DB, userID int64) error {
	res, err := db.Exec(`
        UPDATE users
        SET blocked = TRUE
        WHERE id = $1
    `, userID)
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

func UnblockUser(db *sqlx.DB, userID int64) error {
	res, err := db.Exec(`
        UPDATE users
        SET blocked = FALSE
        WHERE id = $1
    `, userID)
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

func IsUserBlocked(db *sqlx.DB, userID int64) (bool, error) {
	var blocked bool
	err := db.QueryRow(`
        SELECT blocked FROM users WHERE id = $1
    `, userID).Scan(&blocked)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, errors.New("user not found")
		}
		return false, err
	}
	return blocked, nil
}

func CheckUser(db *sqlx.DB, username, passwordHash string) (int64, error) {
	var userID int64
	err := db.QueryRow(`
        SELECT id FROM users
        WHERE username = $1 AND password_hash = $2
    `, username, passwordHash).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return userID, nil
}

// err = db.RestoreUser(db.Postgres, req.Username, req.Mnemonic)
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
	//if userID, _, errFoundUser := GetUserByUsername(db, username); errFoundUser == nil && userID != 0 {
	//	return errors.New("User exists")
	//}
	_, err := db.Exec(`
        INSERT INTO users (username, password_hash, mnemonic, created_at)
        VALUES ($1, $2, $3, $4)
    `, username, passwordHash, mnemonic, time.Now())
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return errors.New("User exists")
			}
		}
		return err
	}
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

// GetUsernameByID returns username by user id (empty if not found)
func GetUsernameByID(db *sqlx.DB, userID int64) (string, error) {
	var username string
	err := db.QueryRow(`
        SELECT username FROM users WHERE id = $1
    `, userID).Scan(&username)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return username, nil
}

func GetAllUserIDs(db *sqlx.DB) ([]int64, error) {
	var userIDs []int64
	err := db.Select(&userIDs, `SELECT id FROM users`)
	if err != nil {
		return nil, err
	}
	return userIDs, nil
}

func IsAdmin(db *sqlx.DB, userID int64) (bool, error) {
	var isAdmin bool
	query := `SELECT is_admin FROM users WHERE id = $1`
	log.Printf("Executing SQL: %s with userID=%d", query, userID)
	err := db.QueryRow(query, userID).Scan(&isAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No user found with id=%d", userID)
			return false, nil
		}
		return false, err
	}
	log.Printf("User %d is_admin=%v", userID, isAdmin)
	return isAdmin, nil
}
func MakeAdmin(db *sqlx.DB, userID int64) error {
	res, err := db.Exec(`
        UPDATE users SET is_admin = TRUE WHERE id = $1
    `, userID)
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

func RemoveAdmin(db *sqlx.DB, userID int64) error {
	res, err := db.Exec(`
        UPDATE users SET is_admin = FALSE WHERE id = $1
    `, userID)
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
