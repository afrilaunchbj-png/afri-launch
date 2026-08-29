package server

import (
	"net/http"
	"sync"
	"time"

	"afrilaunch/backend/internal/server/apierror"
	"afrilaunch/backend/internal/server/handler"
)

// rateLimiter est un limiteur simple en mémoire (fixed window) par IP.
// Suffisant pour le MVP ; en production, utiliser Redis (distribué).
type rateLimiter struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	visitors map[string]*windowEntry
}

type windowEntry struct {
	count int
	reset time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{limit: limit, window: window, visitors: make(map[string]*windowEntry)}
}

// allow enregistre la requête et indique si elle est autorisée.
func (rl *rateLimiter) allow(ip string) bool {
	now := time.Now()
	rl.mu.Lock()
	defer rl.mu.Unlock()

	w, ok := rl.visitors[ip]
	if !ok || now.After(w.reset) {
		rl.visitors[ip] = &windowEntry{count: 1, reset: now.Add(rl.window)}
		return true
	}
	if w.count >= rl.limit {
		return false
	}
	w.count++
	return true
}

// rateLimit applique un plafond par IP (ex. endpoints d'authentification).
func rateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	rl := newRateLimiter(limit, window)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			if !rl.allow(ip) {
				w.Header().Set("Retry-After", "60")
				handler.WriteError(w, r, apierror.TooManyRequests("Trop de requêtes. Réessayez dans un instant."))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}
