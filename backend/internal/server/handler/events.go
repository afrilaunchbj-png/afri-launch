package handler

import (
	"fmt"
	"net/http"
	"time"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/server/authctx"
)

// EventHandler diffuse le canal temps réel unique (SSE) : deltas du chat,
// statuts des jobs, notifications futures. Un flux par utilisateur.
type EventHandler struct {
	bus port.EventBus
}

// NewEventHandler construit le handler du canal temps réel.
func NewEventHandler(bus port.EventBus) *EventHandler { return &EventHandler{bus: bus} }

// Stream gère GET /events (text/event-stream, long-polling SSE).
// Reconnexion : le client resynchronise (refetch) — les données du chat
// sont persistées, aucune n'est perdue définitivement.
func (h *EventHandler) Stream(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	ch, cancel := h.bus.Subscribe(userID)
	defer cancel()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// Proxies : pas de buffering (nécessaire pour le streaming).
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, hasFlusher := w.(http.Flusher)
	write := func(payload string) bool {
		if _, err := fmt.Fprint(w, payload); err != nil {
			return false
		}
		if hasFlusher {
			flusher.Flush()
		}
		return true
	}

	if !write("event: connected\ndata: {}\n\n") {
		return
	}

	// Heartbeat : maintient la connexion traversée par proxies/LB.
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case evt, ok := <-ch:
			if !ok {
				// Abonné déconnecté (buffer plein) : le client se reconnecte.
				return
			}
			if !write("event: " + evt.Type + "\ndata: " + string(evt.Data) + "\n\n") {
				return
			}
		case <-ticker.C:
			if !write(": ping\n\n") {
				return
			}
		case <-r.Context().Done():
			return
		}
	}
}
