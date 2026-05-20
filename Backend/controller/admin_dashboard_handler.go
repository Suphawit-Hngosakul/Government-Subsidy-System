package controller

import (
	"net/http"
	"strconv"

	"government-subsidy-system/backend/service"
)

type AdminDashboardHandler struct {
	service *service.DashboardService
}

func NewAdminDashboardHandler(svc *service.DashboardService) *AdminDashboardHandler {
	return &AdminDashboardHandler{service: svc}
}

func (h *AdminDashboardHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/admin/stats", h.stats)
	mux.HandleFunc("GET /api/v1/admin/audit-log", h.auditLog)
}

func (h *AdminDashboardHandler) stats(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.service.Stats(r.Context()))
}

func (h *AdminDashboardHandler) auditLog(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	writeJSON(w, http.StatusOK, map[string]any{"entries": h.service.AuditLog(r.Context(), limit)})
}
