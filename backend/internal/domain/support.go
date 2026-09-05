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

// SupportAttachment est un fichier joint (capture d'écran, PDF) rattaché à
// un ticket (message initial) ou à un message du fil.
type SupportAttachment struct {
	ID          string
	UserID      string
	TicketID    string
	MessageID   string
	Filename    string
	StorageKey  string
	ContentType string
	SizeBytes   int64
	CreatedAt   time.Time
}

// Pièces jointes : limites de validation.
const (
	AttachmentMaxSize      = 5 << 20 // 5 Mo par fichier
	AttachmentMaxPerSubmit = 4
)

// AttachmentAllowedContentType indique si le type MIME est accepté
// (captures d'écran et PDF).
func AttachmentAllowedContentType(ct string) bool {
	switch ct {
	case "image/png", "image/jpeg", "image/webp", "image/gif", "application/pdf":
		return true
	default:
		return false
	}
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
