package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	"government-subsidy-system/backend/domain"
	"government-subsidy-system/backend/repository"
	"government-subsidy-system/backend/service"
)

type OfficerClaimHandler struct {
	service *service.OfficerService
}

func NewOfficerClaimHandler(svc *service.OfficerService) *OfficerClaimHandler {
	return &OfficerClaimHandler{service: svc}
}

func (h *OfficerClaimHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/officer/claims", h.list)
	mux.HandleFunc("GET /api/v1/officer/claim/{id}", h.detail)
	mux.HandleFunc("PATCH /api/v1/officer/claim/{id}/approve", h.approve)
	mux.HandleFunc("PATCH /api/v1/officer/claim/{id}/reject", h.reject)
}

func (h *OfficerClaimHandler) list(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"claims": h.service.ListPending(r.Context())})
}

func (h *OfficerClaimHandler) detail(w http.ResponseWriter, r *http.Request) {
	detail, err := h.service.Detail(r.Context(), r.PathValue("id"))
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (h *OfficerClaimHandler) approve(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeDecisionInput(w, r)
	if !ok {
		return
	}
	claim, err := h.service.Approve(r.Context(), r.PathValue("id"), input)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, claim)
}

func (h *OfficerClaimHandler) reject(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeDecisionInput(w, r)
	if !ok {
		return
	}
	claim, err := h.service.Reject(r.Context(), r.PathValue("id"), input)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, claim)
}

func decodeDecisionInput(w http.ResponseWriter, r *http.Request) (domain.OfficerDecisionInput, bool) {
	var input domain.OfficerDecisionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request body")
		return domain.OfficerDecisionInput{}, false
	}
	return input, true
}

func (h *OfficerClaimHandler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrOfficerClaimNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrClaimAlreadyDecided):
		writeError(w, http.StatusConflict, err.Error())
	default:
		writeError(w, http.StatusBadRequest, err.Error())
	}
}
