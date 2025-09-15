package db

import (
	"mFrelance/models"

	"github.com/jmoiron/sqlx"
)

// GetTicketByID retrieves a ticket by ID
func GetTicketByID(db *sqlx.DB, id int64) (*models.Ticket, error) {
	return models.GetTicketByID(db, id)
}

// CreateTicket creates a new ticket
func CreateTicket(db *sqlx.DB, subject string, userID int64) (int64, error) {
	return models.CreateTicket(db, subject, userID)
}

// GetRandomOpenTicket retrieves a random open ticket
func GetRandomOpenTicket(db *sqlx.DB) (*models.Ticket, error) {
	return models.GetRandomOpenTicket(db)
}

// AssignTicketAdmin assigns an admin to a ticket
func AssignTicketAdmin(db *sqlx.DB, ticketID, adminID int64) error {
	return models.AssignTicketAdmin(db, ticketID, adminID)
}

// AddUserToTicket adds a user to a ticket
func AddUserToTicket(db *sqlx.DB, ticketID, userID int64) error {
	return models.AddUserToTicket(db, ticketID, userID)
}

// GetMessagesForTicket retrieves messages for a ticket
func GetMessagesForTicket(db *sqlx.DB, ticketID, userID int64) ([]models.TicketMessage, error) {
	return models.GetMessagesForTicket(db, ticketID, userID)
}

// GetTicketsForUser retrieves tickets for a user
func GetTicketsForUser(db *sqlx.DB, userID int64) ([]models.Ticket, error) {
	return models.GetTicketsForUser(db, userID)
}

// ExitFromTicket removes a user from a ticket
func ExitFromTicket(db *sqlx.DB, ticketID, userID int64) error {
	return models.ExitFromTicket(db, ticketID, userID)
}

// CloseTicket closes a ticket
func CloseTicket(db *sqlx.DB, ticketID int64) error {
	return models.CloseTicket(db, ticketID)
}

// OpenTicket opens a ticket
func OpenTicket(db *sqlx.DB, ticketID int64) error {
	return models.OpenTicket(db, ticketID)
}

// PendingTicket sets a ticket to pending
func PendingTicket(db *sqlx.DB, ticketID int64) error {
	return models.PendingTicket(db, ticketID)
}

// AddTicketMessage adds a message to a ticket
func AddTicketMessage(db *sqlx.DB, ticketID, senderID int64, text string) error {
	return models.AddTicketMessage(db, ticketID, senderID, text)
}

// GetTicketMessages retrieves messages for a ticket
func GetTicketMessages(db *sqlx.DB, ticketID int64) ([]models.TicketMessage, error) {
	return models.GetTicketMessages(db, ticketID)
}

// MarkTicketMessagesRead marks ticket messages as read
func MarkTicketMessagesRead(db *sqlx.DB, ticketID int64, userID int64) error {
	return models.MarkTicketMessagesRead(db, ticketID, userID)
}
