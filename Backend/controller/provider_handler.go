package controller

import (
	"errors"
	"net/http"

	"government-subsidy-system/backend/repository"
	"government-subsidy-system/backend/service"
)

type ProviderHandler struct {
	service *service.ProviderService
}

func NewProviderHandler(svc *service.ProviderService) *ProviderHandler {
	return &ProviderHandler{service: svc}
}

func (h *ProviderHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/dopa/verify/{nationalId}", h.verifyDOPA)
	mux.HandleFunc("GET /api/v1/dopa/card-status/{nationalId}", h.dopaCardStatus)
	mux.HandleFunc("GET /api/v1/sso/status/{nationalId}", h.ssoStatus)
	mux.HandleFunc("GET /api/v1/sso/contribution/{nationalId}", h.ssoContribution)
	mux.HandleFunc("GET /api/v1/ktb/financial-check/{nationalId}", h.ktbFinancialCheck)
	mux.HandleFunc("GET /api/v1/ktb/account-status/{nationalId}", h.ktbAccountStatus)
}

func (h *ProviderHandler) verifyDOPA(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.VerifyDOPA(r.Context(), r.PathValue("nationalId"))
	if err != nil {
		handleProviderError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *ProviderHandler) dopaCardStatus(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.DOPACardStatus(r.Context(), r.PathValue("nationalId"))
	if err != nil {
		handleProviderError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *ProviderHandler) ssoStatus(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.SSOStatus(r.Context(), r.PathValue("nationalId"))
	if err != nil {
		handleProviderError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *ProviderHandler) ssoContribution(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.SSOContribution(r.Context(), r.PathValue("nationalId"))
	if err != nil {
		handleProviderError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *ProviderHandler) ktbFinancialCheck(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.KTBFinancialCheck(r.Context(), r.PathValue("nationalId"))
	if err != nil {
		handleProviderError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *ProviderHandler) ktbAccountStatus(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.KTBAccountStatus(r.Context(), r.PathValue("nationalId"))
	if err != nil {
		handleProviderError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func handleProviderError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrNationalIDRequired), errors.Is(err, service.ErrInvalidNationalID):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, repository.ErrProviderRecordNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "provider service unavailable")
	}
}
