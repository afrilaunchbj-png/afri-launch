package domain

import "time"

// AdminProject est un projet vu par le superadmin (avec l'auteur).
type AdminProject struct {
	ID              string
	UserID          string
	Title           string
	Status          string
	CreditsConsumed int64
	CreatedAt       time.Time
	UserEmail       string
	UserName        string
}

// AdminConversation est une conversation vue par le superadmin.
type AdminConversation struct {
	ID        string
	UserID    string
	Title     string
	Status    string
	CreatedAt time.Time
	UserEmail string
	UserName  string
}

// AdminAsset est un asset vu par le superadmin.
type AdminAsset struct {
	ID           string
	ProjectID    string
	Kind         string
	Filename     string
	ContentType  string
	SizeBytes    int64
	CreatedAt    time.Time
	ProjectTitle string
	UserEmail    string
}

// AdminJob est un job de génération vu par le superadmin.
type AdminJob struct {
	ID        string
	UserID    string
	Kind      string
	Status    string
	Cost      int64
	CreatedAt time.Time
	UpdatedAt time.Time
	UserEmail string
	UserName  string
}

// AdminCreditTransaction est une transaction de crédits vue par le superadmin.
type AdminCreditTransaction struct {
	ID        string
	Type      string
	Amount    int64
	Operation string
	Status    string
	CreatedAt time.Time
	UserEmail string
}
