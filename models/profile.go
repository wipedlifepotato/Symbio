package models

import (
    "database/sql/driver"
    "encoding/json"
    "errors"
    "github.com/jmoiron/sqlx"
    "fmt"
    "database/sql"
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
    UserID         int64      `db:"user_id" json:"user_id"`
    FullName       string     `db:"full_name" json:"full_name"`
    Bio            string     `db:"bio" json:"bio"`
    Skills         JSONStrings `db:"skills" json:"skills"`
    Avatar         string     `db:"avatar" json:"avatar"`
    Rating         float64    `db:"rating" json:"rating"`
    CompletedTasks int        `db:"completed_tasks" json:"completed_tasks"`
}

func GetProfile(db *sqlx.DB, userID int64) (*Profile, error) {
    var profile Profile
    err := db.Get(&profile, `SELECT * FROM profiles WHERE user_id=$1`, userID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return &Profile{UserID: userID, Skills: JSONStrings{}}, nil
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

func GetAllProfiles(db *sqlx.DB) ([]Profile, error) {
    var profiles []Profile
    err := db.Select(&profiles, `SELECT * FROM profiles`)
    if err != nil {
        return nil, err
    }
    return profiles, nil
}

