package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"
    "log"

    "github.com/lib/pq"

    "mFrelance/db"
    "mFrelance/models"
    "mFrelance/server"
)

type TicketCreateRequest struct {
    Message string `json:"message"`
    Subject string `json:"subject"`
}

type TicketCreateAnswer struct {
    TicketID int64 `json:"ticket_id"`
}

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
    id, err := models.CreateTicket(server.SanitizeString(t.Subject), claims.UserID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    if err := models.AddTicketMessage(id, claims.UserID, server.SanitizeString(t.Message)); err != nil {
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
    if ticket.UserID != nil && *ticket.UserID == claims.UserID { hasAccess = true }
    if ticket.AdminID != nil && *ticket.AdminID == claims.UserID { hasAccess = true }
    for _, id := range ticket.AdditionalUsers {
        if id == claims.UserID { hasAccess = true; break }
    }
    if !hasAccess {
        server.WriteErrorJSON(w, "you do not have access to this ticket", http.StatusForbidden)
        return
    }
    if err := server.ValidateMessage(req.Message); err != nil {
        server.WriteErrorJSON(w, "invalid message: "+err.Error(), http.StatusBadRequest)
        return
    }
    if err := models.AddTicketMessage(req.TicketID, claims.UserID, req.Message); err != nil {
        server.WriteErrorJSON(w, "failed to add message: "+err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

type TicketIDRequest struct { TicketID int64 `json:"ticket_id"` }

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
    if err := models.ExitFromTicket(req.TicketID, claims.UserID); err != nil {
        server.WriteErrorJSON(w, "failed to exit from ticket: "+err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func GetTicketMessagesHandler(w http.ResponseWriter, r *http.Request) {
    claims := server.GetUserFromContext(r)
    if claims == nil {
        server.WriteErrorJSON(w, "user not found in context", http.StatusUnauthorized)
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
    messages, err := models.GetMessagesForTicket(ticketID, claims.UserID)
    if err != nil {
        server.WriteErrorJSON(w, err.Error(), http.StatusForbidden)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(messages)
}

func GetMyTicketsHandler(w http.ResponseWriter, r *http.Request) {
    claims := server.GetUserFromContext(r)
    if claims == nil {
        server.WriteErrorJSON(w, "user not found in context", http.StatusUnauthorized)
        return
    }
    tickets, err := models.GetTicketsForUser(claims.UserID)
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
    if ticket.UserID != nil && *ticket.UserID == claims.UserID { allowed = true }
    if ticket.AdminID != nil && *ticket.AdminID == claims.UserID { allowed = true }
    for _, id := range ticket.AdditionalUsers {
        if id == claims.UserID { allowed = true; break }
    }
    if !allowed {
        server.WriteErrorJSON(w, "access denied: you are not part of this ticket", http.StatusForbidden)
        return
    }
    if err := models.CloseTicket(req.TicketID); err != nil {
        server.WriteErrorJSON(w, "failed to close ticket: "+err.Error(), http.StatusInternalServerError)
        return
    }
    if err := models.ExitFromTicket(req.TicketID, claims.UserID); err != nil {
        server.WriteErrorJSON(w, "ticket closed, but failed to exit from ticket: "+err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "closed"})
}


