package models

import (
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Ticket struct {
	ID              int64         `db:"id" json:"id"`
	UserID          *int64        `db:"user_id" json:"user_id"`
	AdminID         *int64        `db:"admin_id" json:"admin_id"`
	Status          string        `db:"status" json:"status"`
	Subject         string        `db:"subject" json:"subject"`
	CreatedAt       string        `db:"created_at" json:"created_at"`
	UpdatedAt       string        `db:"updated_at" json:"updated_at"`
	AdditionalUsers pq.Int64Array `db:"additional_users_have_access" json:"additional_users_have_access"`
}

type TicketDoc struct {
	ID              int64   `json:"id"`
	UserID          *int64  `json:"user_id"`
	AdminID         *int64  `json:"admin_id"`
	AdditionalUsers []int64 `json:"additional_users_have_access"`
	Status          string  `json:"status"`
	Subject         string  `json:"subject"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

type TicketMessage struct {
	ID        int64     `db:"id"`
	TicketID  int64     `db:"ticket_id"`
	SenderID  int64     `db:"sender_id"`
	Message   string    `db:"message"`
	Read      bool      `db:"read"`
	CreatedAt time.Time `db:"created_at"`
}

// -----------------------------
// Tickets
// -----------------------------

func GetTicketByID(db *sqlx.DB, id int64) (*Ticket, error) {
	var t Ticket
	err := db.QueryRow(`
		SELECT id, user_id, admin_id, status, subject, created_at, updated_at
		FROM tickets
		WHERE id=$1
	`, id).Scan(&t.ID, &t.UserID, &t.AdminID, &t.Status, &t.Subject, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func CreateTicket(db *sqlx.DB, subject string, userID int64) (int64, error) {
	var id int64
	err := db.QueryRow(`
		INSERT INTO tickets(subject, user_id) 
		VALUES($1, $2) 
		RETURNING id
	`, subject, userID).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
func GetRandomOpenTicket(db *sqlx.DB) (*Ticket, error) {
	var t Ticket
	err := db.QueryRow(`
		SELECT id, user_id, admin_id, status, subject, created_at, updated_at
		FROM tickets
		WHERE status='open' AND admin_id IS NULL
		ORDER BY RANDOM()
		LIMIT 1
	`).Scan(&t.ID, &t.UserID, &t.AdminID, &t.Status, &t.Subject, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func AssignTicketAdmin(db *sqlx.DB, ticketID, adminID int64) error {
	_, err := db.Exec(`
		UPDATE tickets
		SET admin_id=$1, updated_at=NOW()
		WHERE id=$2
	`, adminID, ticketID)
	return err
}

func AddUserToTicket(db *sqlx.DB, ticketID, userID int64) error {
	query := `
        UPDATE tickets
        SET additional_users_have_access = 
            CASE
                WHEN additional_users_have_access IS NULL THEN ARRAY[$1]::INT[]
                WHEN NOT $1 = ANY(additional_users_have_access) THEN additional_users_have_access || $1
                ELSE additional_users_have_access
            END
        WHERE id = $2
    `
	_, err := db.Exec(query, userID, ticketID)
	if err != nil {
		return fmt.Errorf("failed to add user to ticket: %w", err)
	}
	return nil
}

func GetMessagesForTicket(db *sqlx.DB, ticketID, userID int64) ([]TicketMessage, error) {
	var ticket struct {
		UserID          *int64        `db:"user_id"`
		AdminID         *int64        `db:"admin_id"`
		AdditionalUsers pq.Int64Array `db:"additional_users_have_access"`
	}

	err := db.Get(&ticket, `
        SELECT user_id, admin_id, additional_users_have_access
        FROM tickets
        WHERE id=$1
    `, ticketID)
	if err != nil {
		return nil, fmt.Errorf("ticket not found: %w", err)
	}

	hasAccess := false
	if ticket.UserID != nil && *ticket.UserID == userID {
		hasAccess = true
	}
	if ticket.AdminID != nil && *ticket.AdminID == userID {
		hasAccess = true
	}
	for _, id := range ticket.AdditionalUsers {
		if id == userID {
			hasAccess = true
			break
		}
	}
	if !hasAccess {
		return nil, fmt.Errorf("access denied to this ticket")
	}

	var messages []TicketMessage
	err = db.Select(&messages, `
        SELECT *
        FROM ticket_messages
        WHERE ticket_id = $1
        ORDER BY created_at ASC
    `, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return messages, nil
}

func GetTicketsForUser(db *sqlx.DB, userID int64) ([]Ticket, error) {
	var tickets []Ticket
	query := `
        SELECT *
        FROM tickets
        WHERE user_id = $1
           OR admin_id = $1
           OR $1 = ANY(additional_users_have_access)
        ORDER BY created_at DESC
    `
	err := db.Select(&tickets, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tickets: %w", err)
	}
	return tickets, nil
}

func ExitFromTicket(db *sqlx.DB, ticketID, userID int64) error {
	var ticket struct {
		UserID          *int64        `db:"user_id"`
		AdminID         *int64        `db:"admin_id"`
		AdditionalUsers pq.Int64Array `db:"additional_users_have_access"`
	}

	err := db.Get(&ticket, `
        SELECT user_id, admin_id, additional_users_have_access
        FROM tickets
        WHERE id=$1
    `, ticketID)
	if err != nil {
		return fmt.Errorf("ticket not found: %w", err)
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if ticket.UserID != nil && *ticket.UserID == userID {
		_, err := tx.Exec("UPDATE tickets SET user_id = NULL WHERE id=$1", ticketID)
		if err != nil {
			return fmt.Errorf("failed to remove user_id: %w", err)
		}
		return tx.Commit()
	}

	if ticket.AdminID != nil && *ticket.AdminID == userID {
		_, err := tx.Exec("UPDATE tickets SET admin_id = NULL WHERE id=$1", ticketID)
		if err != nil {
			return fmt.Errorf("failed to remove admin_id: %w", err)
		}
		return tx.Commit()
	}

	found := false
	for _, id := range ticket.AdditionalUsers {
		if id == userID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("user is not part of this ticket")
	}

	_, err = tx.Exec(`
        UPDATE tickets
        SET additional_users_have_access = array_remove(additional_users_have_access, $1)
        WHERE id=$2
    `, userID, ticketID)
	if err != nil {
		return fmt.Errorf("failed to remove user from ticket: %w", err)
	}

	return tx.Commit()
}

func CloseTicket(db *sqlx.DB, ticketID int64) error {
	_, err := db.Exec(`
		UPDATE tickets
		SET status='closed', updated_at=NOW()
		WHERE id=$1
	`, ticketID)
	return err
}

func OpenTicket(db *sqlx.DB, ticketID int64) error {
	_, err := db.Exec(`
		UPDATE tickets
		SET status='open', updated_at=NOW()
		WHERE id=$1
	`, ticketID)
	return err
}

func PendingTicket(db *sqlx.DB, ticketID int64) error {
	_, err := db.Exec(`
		UPDATE tickets SET status='pending' WHERE id=$1
	`, ticketID)
	return err
}

// -----------------------------
// Ticket Messages
// -----------------------------

func AddTicketMessage(db *sqlx.DB, ticketID, senderID int64, text string) error {
	_, err := db.Exec(`
		INSERT INTO ticket_messages(ticket_id, sender_id, message, read, created_at)
		VALUES($1, $2, $3, FALSE, NOW())
	`, ticketID, senderID, text)
	return err
}

func GetTicketMessages(db *sqlx.DB, ticketID int64) ([]TicketMessage, error) {
	rows, err := db.Query(`
		SELECT id, ticket_id, sender_id, message, read, created_at
		FROM ticket_messages
		WHERE ticket_id=$1
		ORDER BY created_at ASC
	`, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []TicketMessage
	for rows.Next() {
		var m TicketMessage
		if err := rows.Scan(&m.ID, &m.TicketID, &m.SenderID, &m.Message, &m.Read, &m.CreatedAt); err != nil {
			log.Println("ticket message scan error:", err)
			continue
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func MarkTicketMessagesRead(db *sqlx.DB, ticketID int64, userID int64) error {
	_, err := db.Exec(`
		UPDATE ticket_messages
		SET read=TRUE
		WHERE ticket_id=$1 AND sender_id<>$2
	`, ticketID, userID)
	return err
}
