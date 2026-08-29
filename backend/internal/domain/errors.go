// Package domain contient les entités métier et les erreurs métier.
// Il ne dépend d'aucun framework ni package externe.
package domain

import "errors"

// Erreurs métier sentinelles, typées dans la couche application/server.
var (
	ErrNotFound     = errors.New("resource not found")
	ErrConflict     = errors.New("resource already exists")
	ErrInvalidInput = errors.New("invalid input")
	ErrUnauthorized = errors.New("invalid credentials")
	ErrForbidden    = errors.New("forbidden")
	ErrInsufficient = errors.New("insufficient credits")
	ErrAlreadySaved = errors.New("already saved")
	ErrInvalidToken = errors.New("invalid or expired token")
	ErrNotConfirmed = errors.New("idea not confirmed")
)
