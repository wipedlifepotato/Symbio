package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/lib/pq"

	"mFrelance/db"
	"mFrelance/server"
)

// TicketCreateRequest represents request to create ticket
type TicketCreateRequest struct {
	Message string `json:"message"`
	Subject string `json:"subject"`
}

// TicketCreateAnswer represents response after creating ticket
type TicketCreateAnswer struct {
	TicketID int64 `json:"ticket_id"`
}

// CreateTicket godoc
// @Summary Create new ticket
// @Description Create new ticket
// @Tags ticket
// @Accept  json
// @Produce  json
// @Param request body TicketCreateRequest true "Ticket info"
// @Success 200 {object} TicketCreateAnswer
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/ticket/createTicket [post]
// @Security bearerAuth
func CreateTicket(w http.ResponseWriter, r *http.Request) {
	claims := server.GetUserFromContext(r)
	if claims == nil {
		server.WriteErrorJSON(w, "user not found in context", http.StatusUnauthorized)
		return
	}
	var t TicketCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		log.Println("Decode error:", err)
		http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := server.ValidateTicketField(t.Subject, t.Message); err != nil {
		server.WriteErrorJSON(w, "invalid parameters for ticket: "+err.Error(), http.StatusBadRequest)
		return
	}
	id, err := db.CreateTicket(db.Postgres, server.SanitizeString(t.Subject), claims.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := db.AddTicketMessage(db.Postgres, id, claims.UserID, server.SanitizeString(t.Message)); err != nil {
		server.WriteErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TicketCreateAnswer{TicketID: id})
}

type WriteTicketRequest struct {
	TicketID int64  `json:"ticket_id"`
	Message  string `json:"message"`
}

// WriteToTicketHandler godoc
// @Summary Write to ticket
// @Description Add message to ticket
// @Tags ticket
// @Accept  json
// @Produce  json
// @Param request body WriteTicketRequest true "Message info"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/ticket/write [post]
// @Security bearerAuth
func WriteToTicketHandler(w http.ResponseWriter, r *http.Request) {
	claims := server.GetUserFromContext(r)
	if claims == nil {
		server.WriteErrorJSON(w, "user not found in context", http.StatusUnauthorized)
		return
	}
	var req WriteTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		server.WriteErrorJSON(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	var ticket struct {
		UserID          *int64        `db:"user_id"`
		AdminID         *int64        `db:"admin_id"`
		AdditionalUsers pq.Int64Array `db:"additional_users_have_access"`
	}
	err := db.Postgres.Get(&ticket, `
        SELECT user_id, admin_id, additional_users_have_access
        FROM tickets
        WHERE id=$1
    `, req.TicketID)
	if err != nil {
		server.WriteErrorJSON(w, "ticket not found: "+err.Error(), http.StatusNotFound)
		return
	}
	hasAccess := false
	if ticket.UserID != nil && *ticket.UserID == claims.UserID {
		hasAccess = true
	}
	if ticket.AdminID != nil && *ticket.AdminID == claims.UserID {
		hasAccess = true
	}
	for _, id := range ticket.AdditionalUsers {
		if id == claims.UserID {
			hasAccess = true
			break
		}
	}
	if !hasAccess {
		server.WriteErrorJSON(w, "you do not have access to this ticket", http.StatusForbidden)
		return
	}
	if err := server.ValidateMessage(req.Message); err != nil {
		server.WriteErrorJSON(w, "invalid message: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := db.AddTicketMessage(db.Postgres, req.TicketID, claims.UserID, req.Message); err != nil {
		server.WriteErrorJSON(w, "failed to add message: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

type TicketIDRequest struct {
	TicketID int64 `json:"ticket_id"`
}

// ExitFromTicketHandler godoc
// @Summary Exit from a ticket
// @Description Removes the user from the ticket's participants
// @Tags ticket
// @Accept json
// @Produce json
// @Param request body TicketIDRequest true "Ticket ID"
// @Success 200 {object} map[string]string "status: ok"
// @Failure 400 {object} map[string]string "Invalid payload or ticket_id"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/ticket/exit [post]
// @Security bearerAuth
func ExitFromTicketHandler(w http.ResponseWriter, r *http.Request) {
	claims := server.GetUserFromContext(r)
	if claims == nil {
		server.WriteErrorJSON(w, "user not found in context", http.StatusUnauthorized)
		return
	}
	var req TicketIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		server.WriteErrorJSON(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.TicketID == 0 {
		server.WriteErrorJSON(w, "ticket_id is required", http.StatusBadRequest)
		return
	}
	if err := db.ExitFromTicket(db.Postgres, req.TicketID, claims.UserID); err != nil {
		server.WriteErrorJSON(w, "failed to exit from ticket: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// GetTicketMessagesHandler godoc
// @Summary Get messages for a ticket
// @Description Returns messages for a given ticket if the user has access, supports limit/offset
// @Tags ticket
// @Produce json
// @Param ticket_id query int true "Ticket ID"
// @Param limit query int false "Number of messages to return (max 1000, default 100)"
// @Param offset query int false "Offset for messages (default last messages)"
// @Success 200 {array} models.TicketMessage "List of messages"
// @Failure 400 {object} map[string]string "Invalid parameters"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 403 {object} map[string]string "User does not have access"
// @Router /api/ticket/messages [get]
// @Security bearerAuth
func GetTicketMessagesHandler(w http.ResponseWriter, r *http.Request) {
	claims := server.GetUserFromContext(r)
	if claims == nil {
		server.WriteErrorJSON(w, "user not authenticated", http.StatusUnauthorized)
		return
	}

	ticketIDStr := r.URL.Query().Get("ticket_id")
	if ticketIDStr == "" {
		server.WriteErrorJSON(w, "ticket_id is required", http.StatusBadRequest)
		return
	}
	ticketID, err := strconv.ParseInt(ticketIDStr, 10, 64)
	if err != nil {
		server.WriteErrorJSON(w, "invalid ticket_id", http.StatusBadRequest)
		return
	}

	limit := 100
	offset := 0

	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		l, err := strconv.Atoi(lStr)
		if err != nil || l <= 0 {
			server.WriteErrorJSON(w, "invalid limit", http.StatusBadRequest)
			return
		}
		if l > 1000 {
			server.WriteErrorJSON(w, "limit cannot exceed 1000", http.StatusBadRequest)
			return
		}
		limit = l
	}

	if oStr := r.URL.Query().Get("offset"); oStr != "" {
		o, err := strconv.Atoi(oStr)
		if err != nil || o < 0 {
			server.WriteErrorJSON(w, "invalid offset", http.StatusBadRequest)
			return
		}
		offset = o
	}

	messages, err := db.GetMessagesForTicket(db.Postgres, ticketID, claims.UserID, limit, offset)
	if err != nil {
		server.WriteErrorJSON(w, err.Error(), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// GetMyTicketsHandler godoc
// @Summary Get own tickets
// @Description Get all tickets of user
// @Tags ticket
// @Produce json
// @Success 200 {array} models.TicketDoc
// @Failure 401 {object} map[string]string
// @Router /api/ticket/my [get]
// @Security bearerAuth
func GetMyTicketsHandler(w http.ResponseWriter, r *http.Request) {
	claims := server.GetUserFromContext(r)
	if claims == nil {
		server.WriteErrorJSON(w, "user not found in context", http.StatusUnauthorized)
		return
	}
	tickets, err := db.GetTicketsForUser(db.Postgres, claims.UserID)
	if err != nil {
		server.WriteErrorJSON(w, "failed to get tickets: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tickets)
}

func CloseTicketHandler(w http.ResponseWriter, r *http.Request) {
	claims := server.GetUserFromContext(r)
	if claims == nil {
		server.WriteErrorJSON(w, "user not found in context", http.StatusUnauthorized)
		return
	}
	var req TicketIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		server.WriteErrorJSON(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.TicketID == 0 {
		server.WriteErrorJSON(w, "ticket_id is required", http.StatusBadRequest)
		return
	}
	var ticket struct {
		UserID          *int64        `db:"user_id"`
		AdminID         *int64        `db:"admin_id"`
		AdditionalUsers pq.Int64Array `db:"additional_users_have_access"`
	}
	err := db.Postgres.Get(&ticket, `
        SELECT user_id, admin_id, additional_users_have_access
        FROM tickets
        WHERE id=$1
    `, req.TicketID)
	if err != nil {
		server.WriteErrorJSON(w, "ticket not found: "+err.Error(), http.StatusNotFound)
		return
	}
	allowed := false
	if ticket.UserID != nil && *ticket.UserID == claims.UserID {
		allowed = true
	}
	if ticket.AdminID != nil && *ticket.AdminID == claims.UserID {
		allowed = true
	}
	for _, id := range ticket.AdditionalUsers {
		if id == claims.UserID {
			allowed = true
			break
		}
	}
	if !allowed {
		server.WriteErrorJSON(w, "access denied: you are not part of this ticket", http.StatusForbidden)
		return
	}
	if err := db.CloseTicket(db.Postgres, req.TicketID); err != nil {
		server.WriteErrorJSON(w, "failed to close ticket: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := db.ExitFromTicket(db.Postgres, req.TicketID, claims.UserID); err != nil {
		server.WriteErrorJSON(w, "ticket closed, but failed to exit from ticket: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "closed"})
}
