package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type JSONStrings []string

func (j *JSONStrings) Scan(value interface{}) error {
	if value == nil {
		*j = []string{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("expected []byte, got %T", value)
	}

	return json.Unmarshal(bytes, j)
}

func (j JSONStrings) Value() (driver.Value, error) {
	return json.Marshal(j)
}

type Profile struct {
	UserID         int64       `db:"user_id" json:"user_id"`
	FullName       string      `db:"full_name" json:"full_name"`
	Bio            string      `db:"bio" json:"bio"`
	Skills         JSONStrings `db:"skills" json:"skills"`
	Avatar         string      `db:"avatar" json:"avatar"`
	Rating         float64     `db:"rating" json:"rating"`
	CompletedTasks int         `db:"completed_tasks" json:"completed_tasks"`
	IsAdmin        bool        `db:"is_admin" json:"is_admin"`
	AdminTitle     string      `db:"admin_title" json:"admin_title"`
	Permissions    int         `db:"permissions" json:"permissions"`
}

func (j *JSONStrings) UnmarshalJSON(data []byte) error {
	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	*j = arr
	return nil
}

func GetProfile(db *sqlx.DB, userID int64) (*Profile, error) {
	var profile Profile
	err := db.Get(&profile, `
		SELECT p.*, u.is_admin, COALESCE(u.admin_title, '') as admin_title, u.permissions
		FROM profiles p
		LEFT JOIN users u ON p.user_id = u.id
		WHERE p.user_id=$1
	`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Get user data even if profile doesn't exist
			var isAdmin bool
			var adminTitle string
			var permissions int
			err := db.QueryRow(`
				SELECT is_admin, COALESCE(admin_title, ''), permissions
				FROM users WHERE id=$1
			`, userID).Scan(&isAdmin, &adminTitle, &permissions)
			if err != nil {
				return nil, err
			}
			return &Profile{
				UserID: userID,
				Skills: JSONStrings{},
				IsAdmin: isAdmin,
				AdminTitle: adminTitle,
				Permissions: permissions,
			}, nil
		}
		return nil, err
	}
	return &profile, nil
}

func UpsertProfile(db *sqlx.DB, p *Profile) error {
	_, err := db.Exec(`
        INSERT INTO profiles(user_id, full_name, bio, skills, avatar)
        VALUES($1, $2, $3, $4, $5)
        ON CONFLICT (user_id)
        DO UPDATE SET full_name=$2, bio=$3, skills=$4, avatar=$5
    `, p.UserID, p.FullName, p.Bio, p.Skills, p.Avatar)
	return err
}

func GetProfilesWithLimitOffset(db *sqlx.DB, limit, offset int) ([]Profile, error) {
	var profiles []Profile
	err := db.Select(&profiles, `
		SELECT p.*, u.is_admin, COALESCE(u.admin_title, '') as admin_title, u.permissions
		FROM profiles p
		LEFT JOIN users u ON p.user_id = u.id
		ORDER BY u.is_admin DESC, p.rating DESC, p.completed_tasks DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	return profiles, nil
}
