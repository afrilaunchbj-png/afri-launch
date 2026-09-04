package handler

import (
	"net/http"

	"afrilaunch/backend/internal/application/dashboard"
	"afrilaunch/backend/internal/server/authctx"
)

// DashboardHandler expose les indicateurs personnels du tableau de bord.
type DashboardHandler struct {
	svc *dashboard.Service
}

// NewDashboardHandler construit le handler du tableau de bord.
func NewDashboardHandler(svc *dashboard.Service) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

// Stats gère GET /dashboard/stats : compteurs et séries temporelles de
// l'utilisateur connecté.
func (h *DashboardHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.Stats(r.Context(), authctx.UserID(r.Context()))
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, stats)
}
