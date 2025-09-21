package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type FlexibleTime struct {
	time.Time
}

func (ft *FlexibleTime) Scan(value interface{}) error {
	if value == nil {
		ft.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		ft.Time = v
		return nil
	case []byte:
		parsed, err := time.Parse(time.RFC3339, string(v))
		if err != nil {
			return err
		}
		ft.Time = parsed
		return nil
	case string:
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return err
		}
		ft.Time = parsed
		return nil
	default:
		return fmt.Errorf("cannot scan type %T into FlexibleTime", value)
	}
}

func (ft FlexibleTime) Value() (driver.Value, error) {
	// Если нулевое время → NULL
	if ft.Time.IsZero() {
		return nil, nil
	}
	return ft.Time, nil
}

func (ft *FlexibleTime) UnmarshalJSON(data []byte) error {

	s := strings.Trim(string(data), `"`)

	formats := []string{
		"2006-01-02T15:04:05Z07:00", // RFC3339
		"2006-01-02T15:04:05Z",      // RFC3339 без timezone
		"2006-01-02T15:04:05",       // Без timezone
		"2006-01-02T15:04",          // Без секунд
		"2006-01-02 15:04:05",       // С пробелом
		"2006-01-02 15:04",          // С пробелом без секунд
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			ft.Time = t
			return nil
		}
	}

	return &time.ParseError{
		Layout:     "multiple formats tried",
		Value:      s,
		LayoutElem: "time",
		ValueElem:  s,
	}
}

func (ft FlexibleTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(ft.Time.Format(time.RFC3339))
}

type Task struct {
	ID          int64        `db:"id" json:"id"`
	ClientID    int64        `db:"client_id" json:"client_id"`
	Title       string       `db:"title" json:"title"`
	Description string       `db:"description" json:"description"`
	Category    string       `db:"category" json:"category"`
	Budget      float64      `db:"budget" json:"budget"`
	Currency    string       `db:"currency" json:"currency"`
	Status      string       `db:"status" json:"status"` // open, in_progress, completed, cancelled, disputed
	CreatedAt   time.Time    `db:"created_at" json:"created_at"`
	Deadline    FlexibleTime `db:"deadline" json:"deadline"`
}
