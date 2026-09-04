package domain

import "time"

// SupportTicket est une demande d'assistance d'un utilisateur.
type SupportTicket struct {
	ID        string
	UserID    string
	Subject   string
	Message   string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AdminTicket est un ticket vu par le superadmin (avec l'auteur).
type AdminTicket struct {
	SupportTicket
	UserEmail string
	UserName  string
}

// TicketMessage est un message du fil de discussion d'un ticket.
type TicketMessage struct {
	ID        string
	TicketID  string
	AuthorID  string
	Content   string
	IsAdmin   bool
	CreatedAt time.Time
}

// TicketMessageView est un message enrichi des informations de l'auteur.
type TicketMessageView struct {
	TicketMessage
	AuthorEmail string
	AuthorName  string
}

// Statuts d'un ticket de support.
const (
	TicketOpen     = "open"
	TicketResolved = "resolved"
)

// AdminStats regroupe les indicateurs du suivi global superadmin.
type AdminStats struct {
	Users           int64            `json:"users"`
	Projects        int64            `json:"projects"`
	Assets          int64            `json:"assets"`
	Conversations   int64            `json:"conversations"`
	JobsByStatus    map[string]int64 `json:"jobs_by_status"`
	CreditsConsumed int64            `json:"credits_consumed"`
	OpenTickets     int64            `json:"open_tickets"`
}
