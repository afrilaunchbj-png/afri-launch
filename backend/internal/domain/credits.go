package domain

import "time"

// Types de mouvement du ledger de crédits.
const (
	TransactionCredit = "credit"
	TransactionDebit  = "debit"
)

// Opérations métier facturables (coût configurable via generation_costs).
const (
	OperationWelcomeBonus   = "welcome_bonus"
	OperationPurchase       = "purchase"
	OperationRefund         = "refund"
	OperationNicheResearch  = "niche_research"
	OperationIdeaGeneration = "idea_generation"
	OperationEbookGen       = "ebook_generation"
	OperationImageGen       = "image_generation"
	OperationPosterGen      = "poster_generation"
	OperationVideoGen       = "video_generation"
	OperationSalesPage      = "sales_page"
)

// Statuts d'une réservation.
const (
	ReservationReserved = "reserved"
	ReservationConsumed = "consumed"
	ReservationReleased = "released"
)

// CreditAccount est le solde d'un utilisateur. `available = balance - reserved`.
type CreditAccount struct {
	ID        string
	UserID    string
	Balance   int64
	Reserved  int64
	Version   int32
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Available renvoie les crédits réellement disponibles.
func (a CreditAccount) Available() int64 {
	return a.Balance - a.Reserved
}

// CreditTransaction est une écriture du journal comptable (append-only).
type CreditTransaction struct {
	ID        string
	AccountID string
	Type      string
	Amount    int64
	Operation string
	Status    string
	Reference *string
	Metadata  map[string]any
	CreatedAt time.Time
}

// CreditReservation bloque temporairement des crédits avant exécution.
type CreditReservation struct {
	ID        string
	AccountID string
	Amount    int64
	Operation string
	Reference string
	Status    string
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// GenerationCost définit le coût configurable d'une opération.
type GenerationCost struct {
	Operation string `json:"operation"`
	Name      string `json:"name"`
	Credits   int64  `json:"credits"`
	IsActive  bool   `json:"is_active"`
}

// CreditSummary agrège les métriques d'un compte pour l'affichage.
type CreditSummary struct {
	Balance    int64 `json:"balance"`
	Reserved   int64 `json:"reserved"`
	Available  int64 `json:"available"`
	AddedMonth int64 `json:"added_month"`
	UsedMonth  int64 `json:"used_month"`
}
