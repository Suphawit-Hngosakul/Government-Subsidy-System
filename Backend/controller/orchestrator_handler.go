package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"government-subsidy-system/backend/domain"
	"government-subsidy-system/backend/service"
)


type OrchestratorHTTPHandler struct {
	service *service.OrchestratorService
}

func NewOrchestratorHTTPHandler(service *service.OrchestratorService) *OrchestratorHTTPHandler {
	return &OrchestratorHTTPHandler{service: service}
}

func (h *OrchestratorHTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("POST /internal/orchestrate", h.orchestrate)
	mux.HandleFunc("POST /internal/decision", h.decision)
	mux.HandleFunc("GET /api/v1/claim/", h.streamClaimStatus)
}

func (h *OrchestratorHTTPHandler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *OrchestratorHTTPHandler) orchestrate(w http.ResponseWriter, r *http.Request) {
	var req domain.OrchestrateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}

	result, err := h.service.Orchestrate(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, result)
}

func (h *OrchestratorHTTPHandler) decision(w http.ResponseWriter, r *http.Request) {
	var req domain.DecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}

	result, err := h.service.Decision(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *OrchestratorHTTPHandler) streamClaimStatus(w http.ResponseWriter, r *http.Request) {
	claimID, ok := claimIDFromStreamPath(r.URL.Path)
	if !ok {
		writeError(w, http.StatusNotFound, "stream endpoint not found")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming is not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	for _, event := range h.service.Events(claimID) {
		writeSSE(w, event)
	}
	flusher.Flush()

	events, unsubscribe := h.service.Subscribe(claimID)
	defer unsubscribe()

	for {
		select {
		case <-r.Context().Done():
			return
		case event := <-events:
			writeSSE(w, event)
			flusher.Flush()
		}
	}
}

func claimIDFromStreamPath(path string) (string, bool) {
	const prefix = "/api/v1/claim/"
	const suffix = "/stream"
	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, suffix) {
		return "", false
	}

	claimID := strings.TrimSuffix(strings.TrimPrefix(path, prefix), suffix)
	return claimID, claimID != "" && !strings.Contains(claimID, "/")
}

func writeSSE(w http.ResponseWriter, event domain.StatusEvent) {
	payload, _ := json.Marshal(event)
	fmt.Fprintf(w, "event: claim-status\n")
	fmt.Fprintf(w, "data: %s\n\n", payload)
}
