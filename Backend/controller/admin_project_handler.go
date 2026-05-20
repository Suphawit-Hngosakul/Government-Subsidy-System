package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	"government-subsidy-system/backend/domain"
	"government-subsidy-system/backend/repository"
	"government-subsidy-system/backend/service"
)

type AdminProjectHandler struct {
	service *service.ProjectService
}

func NewAdminProjectHandler(svc *service.ProjectService) *AdminProjectHandler {
	return &AdminProjectHandler{service: svc}
}

func (h *AdminProjectHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/admin/projects", h.list)
	mux.HandleFunc("POST /api/v1/admin/project", h.create)
	mux.HandleFunc("GET /api/v1/admin/project/{id}", h.get)
	mux.HandleFunc("PUT /api/v1/admin/project/{id}", h.update)
	mux.HandleFunc("DELETE /api/v1/admin/project/{id}", h.remove)
}

func (h *AdminProjectHandler) list(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"projects": h.service.List(r.Context())})
}

func (h *AdminProjectHandler) create(w http.ResponseWriter, r *http.Request) {
	var input domain.ProjectInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}

	project, err := h.service.Create(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, project)
}

func (h *AdminProjectHandler) get(w http.ResponseWriter, r *http.Request) {
	project, err := h.service.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func (h *AdminProjectHandler) update(w http.ResponseWriter, r *http.Request) {
	var patch domain.ProjectUpdate
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}

	project, err := h.service.Update(r.Context(), r.PathValue("id"), patch)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func (h *AdminProjectHandler) remove(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.Context(), r.PathValue("id")); err != nil {
		h.handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminProjectHandler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrProjectNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		writeError(w, http.StatusBadRequest, err.Error())
	}
}
