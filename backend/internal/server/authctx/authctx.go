// Package authctx transporte l'identité utilisateur (issue de Neon Auth)
// dans le contexte HTTP.
package authctx

import (
	"context"

	"afrilaunch/backend/internal/application/port"
)

type key int

const userKey key = iota

// WithUser attache l'identité vérifiée au contexte.
func WithUser(ctx context.Context, u port.AuthUser) context.Context {
	return context.WithValue(ctx, userKey, u)
}

// User récupère l'identité utilisateur, le cas échéant.
func User(ctx context.Context) (port.AuthUser, bool) {
	u, ok := ctx.Value(userKey).(port.AuthUser)
	return u, ok && u.ID != ""
}

// UserID renvoie l'identifiant utilisateur courant (ou "").
func UserID(ctx context.Context) string {
	u, _ := User(ctx)
	return u.ID
}
