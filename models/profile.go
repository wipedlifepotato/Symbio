package models

import (
   // "errors"
    "github.com/jmoiron/sqlx"
    //"log"
    "database/sql"
    "encoding/json"
    "errors"
)


type Profile struct {
    UserID        int64    `db:"user_id" json:"user_id"`
    FullName      string   `db:"full_name" json:"full_name"`
    Bio           string   `db:"bio" json:"bio"`
    Skills        []string `db:"skills" json:"skills"`
    Avatar        string   `db:"avatar" json:"avatar"`
    Rating        float64  `db:"rating" json:"rating"`
    CompletedTasks int     `db:"completed_tasks" json:"completed_tasks"`
}

//func (p *Profile) GetProfile(db

func GetProfile(db *sqlx.DB, userID int64) (*Profile, error) {
    var profile struct {
        UserID        int64          `db:"user_id"`
        FullName      string         `db:"full_name"`
        Bio           string         `db:"bio"`
        SkillsRaw     []byte         `db:"skills"`
        Avatar        string         `db:"avatar"`
        Rating        float64        `db:"rating"`
        CompletedTasks int           `db:"completed_tasks"`
    }

    err := db.Get(&profile, `SELECT * FROM profiles WHERE user_id=$1`, userID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return &Profile{UserID: userID}, nil
        }
        return nil, err
    }

    var skills []string
    if len(profile.SkillsRaw) > 0 {
        if err := json.Unmarshal(profile.SkillsRaw, &skills); err != nil {
            return nil, err
        }
    }

    return &Profile{
        UserID: profile.UserID,
        FullName: profile.FullName,
        Bio: profile.Bio,
        Skills: skills,
        Avatar: profile.Avatar,
        Rating: profile.Rating,
        CompletedTasks: profile.CompletedTasks,
    }, nil
}
func UpsertProfile(db *sqlx.DB, p *Profile) error {
    skillsJSON, err := json.Marshal(p.Skills)
    if err != nil {
        return err
    }

    _, err = db.Exec(`
        INSERT INTO profiles(user_id, full_name, bio, skills, avatar)
        VALUES($1, $2, $3, $4, $5)
        ON CONFLICT (user_id)
        DO UPDATE SET full_name=$2, bio=$3, skills=$4, avatar=$5
    `, p.UserID, p.FullName, p.Bio, skillsJSON, p.Avatar)
    return err
}

