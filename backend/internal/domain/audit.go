package domain

import "time"

// AuditLog est une entrée du journal des opérations sensibles (append-only).
type AuditLog struct {
	ID        string
	UserID    string
	Action    string
	Entity    string
	EntityID  string
	Metadata  map[string]any
	CreatedAt time.Time
}

// Actions auditables (journal d'activités de l'administration).
const (
	AuditUserRegister     = "user.register"
	AuditUserRolePromoted = "user.role_promoted"
	AuditTicketCreate     = "ticket.create"
	AuditTicketReply      = "ticket.reply"
	AuditTicketAdminReply = "ticket.admin_reply"
	AuditTicketResolve    = "ticket.resolve"

	// Générations (jobs) et intégrations publicitaires.
	AuditGenerationDispatched   = "generation.dispatched"
	AuditGenerationCompleted    = "generation.completed"
	AuditGenerationFailed       = "generation.failed"
	AuditProjectCreated         = "project.created"
	AuditCampaignCreated        = "campaign.created"
	AuditCampaignStatusChanged  = "campaign.status_changed"
	AuditCreativePublished      = "creative.published"
	AuditConnectionDisconnected = "connection.disconnected"

	// Paiements (recharges de crédits).
	AuditPaymentCheckout  = "payment.checkout"
	AuditPaymentSucceeded = "payment.succeeded"
	AuditPaymentFailed    = "payment.failed"
)
