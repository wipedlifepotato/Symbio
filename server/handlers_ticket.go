package server

import (
    "encoding/json"
    "log"
    "mFrelance/db"
    "net/http"
    "strconv"
    "mFrelance/models"
)
import "github.com/lib/pq"
//import "io"

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
	claims := GetUserFromContext(r)
	if claims == nil {
		writeErrorJSON(w, "user not found in context", http.StatusUnauthorized)
		return
	}

	var t TicketCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		log.Println("Decode error:", err)
		http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	err := ValidateTicketField(t.Subject,t.Message)
	if err != nil {
		writeErrorJSON(w, "invalid parameters for ticket: "+err.Error(), http.StatusBadRequest)
		return
	}

	id, err := models.CreateTicket(SanitizeString(t.Subject), claims.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = models.AddTicketMessage(id, claims.UserID, SanitizeString(t.Message))
	if err != nil {
		writeErrorJSON(w, err.Error(), http.StatusBadRequest)
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
    claims := GetUserFromContext(r)
    if claims == nil {
        writeErrorJSON(w, "user not found in context", http.StatusUnauthorized)
        return
    }
    var req WriteTicketRequest

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeErrorJSON(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
        return
    }

    var ticket struct {
        UserID          *int64  `db:"user_id"`
        AdminID         *int64  `db:"admin_id"`
        AdditionalUsers pq.Int64Array `db:"additional_users_have_access"`
    }

    err := db.Postgres.Get(&ticket, `
        SELECT user_id, admin_id, additional_users_have_access
        FROM tickets
        WHERE id=$1
    `, req.TicketID)
    if err != nil {
        writeErrorJSON(w, "ticket not found: "+err.Error(), http.StatusNotFound)
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
        writeErrorJSON(w, "you do not have access to this ticket", http.StatusForbidden)
        return
    }

    if err := ValidateMessage(req.Message); err != nil {
        writeErrorJSON(w, "invalid message: "+err.Error(), http.StatusBadRequest)
        return
    }

    err = models.AddTicketMessage(req.TicketID, claims.UserID, req.Message)
    if err != nil {
        writeErrorJSON(w, "failed to add message: "+err.Error(), http.StatusInternalServerError)
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
    claims := GetUserFromContext(r)
    if claims == nil {
        writeErrorJSON(w, "user not found in context", http.StatusUnauthorized)
        return
    }

    var req struct {
        TicketID int64 `json:"ticket_id"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeErrorJSON(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
        return
    }

    if req.TicketID == 0 {
        writeErrorJSON(w, "ticket_id is required", http.StatusBadRequest)
        return
    }

    err := models.ExitFromTicket(req.TicketID, claims.UserID)
    if err != nil {
        writeErrorJSON(w, "failed to exit from ticket: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
// GetTicketMessagesHandler godoc
// @Summary Get messages for a ticket
// @Description Returns all messages for a given ticket if the user has access
// @Tags ticket
// @Produce json
// @Param ticket_id query int true "Ticket ID"
// @Success 200 {array} models.TicketMessage "List of messages"
// @Failure 400 {object} map[string]string "Invalid ticket_id"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 403 {object} map[string]string "User does not have access"
// @Router /api/ticket/messages [get]
// @Security bearerAuth
func GetTicketMessagesHandler(w http.ResponseWriter, r *http.Request) {
    claims := GetUserFromContext(r)
    if claims == nil {
        writeErrorJSON(w, "user not found in context", http.StatusUnauthorized)
        return
    }

    ticketIDStr := r.URL.Query().Get("ticket_id")
    if ticketIDStr == "" {
        writeErrorJSON(w, "ticket_id is required", http.StatusBadRequest)
        return
    }

    ticketID, err := strconv.ParseInt(ticketIDStr, 10, 64)
    if err != nil {
        writeErrorJSON(w, "invalid ticket_id", http.StatusBadRequest)
        return
    }

    messages, err := models.GetMessagesForTicket(ticketID, claims.UserID)
    if err != nil {
        writeErrorJSON(w, err.Error(), http.StatusForbidden)
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
    claims := GetUserFromContext(r)
    if claims == nil {
        writeErrorJSON(w, "user not found in context", http.StatusUnauthorized)
        return
    }

    tickets, err := models.GetTicketsForUser(claims.UserID)
    if err != nil {
        writeErrorJSON(w, "failed to get tickets: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(tickets)
}

func CloseTicketHandler(w http.ResponseWriter, r *http.Request) {
    claims := GetUserFromContext(r)
    if claims == nil {
        writeErrorJSON(w, "user not found in context", http.StatusUnauthorized)
        return
    }

    var req struct {
        TicketID int64 `json:"ticket_id"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeErrorJSON(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
        return
    }

    if req.TicketID == 0 {
        writeErrorJSON(w, "ticket_id is required", http.StatusBadRequest)
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
        writeErrorJSON(w, "ticket not found: "+err.Error(), http.StatusNotFound)
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
        writeErrorJSON(w, "access denied: you are not part of this ticket", http.StatusForbidden)
        return
    }

    if err := models.CloseTicket(req.TicketID); err != nil {
        writeErrorJSON(w, "failed to close ticket: "+err.Error(), http.StatusInternalServerError)
        return
    }

    if err := models.ExitFromTicket(req.TicketID, claims.UserID); err != nil {
        writeErrorJSON(w, "ticket closed, but failed to exit from ticket: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "closed"})
}