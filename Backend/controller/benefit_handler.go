package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"government-subsidy-system/backend/domain"
	"government-subsidy-system/backend/repository"
	"government-subsidy-system/backend/service"
)

type BenefitHandler struct {
	service *service.BenefitService
}

func NewBenefitHandler(svc *service.BenefitService) *BenefitHandler {
	return &BenefitHandler{service: svc}
}

func (h *BenefitHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/benefit/claim", h.submitClaim)
	mux.HandleFunc("GET /api/v1/benefit/status/", h.getStatus)
	mux.HandleFunc("GET /api/v1/benefit/history/", h.getHistory)
	mux.HandleFunc("GET /api/v1/benefit/projects", h.getProjects)
}

// POST /api/v1/benefit/claim - ยื่นคำร้องขอสิทธิ์
func (h *BenefitHandler) submitClaim(w http.ResponseWriter, r *http.Request) {
	var req domain.BenefitClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}

	resp, err := h.service.SubmitClaim(r.Context(), req)
	if err != nil {
		if err.Error() == "project not found" {
			writeError(w, http.StatusNotFound, "project not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// GET /api/v1/benefit/status/:claimId - เช็คสถานะ real-time
func (h *BenefitHandler) getStatus(w http.ResponseWriter, r *http.Request) {
	claimID, ok := extractPathParam(r.URL.Path, "/api/v1/benefit/status/")
	if !ok || claimID == "" {
		writeError(w, http.StatusBadRequest, "claimId is required")
		return
	}

	resp, err := h.service.GetClaimStatus(r.Context(), claimID)
	if err != nil {
		if errors.Is(err, repository.ErrClaimNotFound) {
			writeError(w, http.StatusNotFound, "claim not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// GET /api/v1/benefit/history/:citizenId - ประวัติคำร้องทั้งหมด
func (h *BenefitHandler) getHistory(w http.ResponseWriter, r *http.Request) {
	nationalID, ok := extractPathParam(r.URL.Path, "/api/v1/benefit/history/")
	if !ok || nationalID == "" {
		writeError(w, http.StatusBadRequest, "nationalId is required")
		return
	}

	claims, err := h.service.GetClaimHistory(r.Context(), nationalID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"nationalId": nationalID,
		"claims":     claims,
	})
}

// GET /api/v1/benefit/projects - รายการโครงการที่เปิดรับสิทธิ์
func (h *BenefitHandler) getProjects(w http.ResponseWriter, r *http.Request) {
	projects := h.service.GetAvailableProjects(r.Context())

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"projects": projects,
	})
}

func extractPathParam(path string, prefix string) (string, bool) {
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}

	param := strings.TrimPrefix(path, prefix)
	// Stop at query string or trailing slash
	if idx := strings.IndexAny(param, "?/"); idx != -1 {
		param = param[:idx]
	}

	return param, param != ""
}
