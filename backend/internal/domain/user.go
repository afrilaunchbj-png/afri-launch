package domain

import "time"

// User représente un compte utilisateur. Avec Neon Auth (Managed Better Auth),
// l'identité (email, mot de passe, sessions) est gérée par Neon : notre table
// `users` est une table « profil » cléée par le `sub` (ID utilisateur Neon).
type User struct {
	ID        string
	Email     string
	FullName  string
	AvatarURL *string
	CreatedAt time.Time
}
