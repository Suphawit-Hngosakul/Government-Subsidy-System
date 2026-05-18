package controller

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"government-subsidy-system/backend/domain"
	"government-subsidy-system/backend/service"
)

type AuthHandler struct {
	auth service.AuthService
}

func NewAuthHandler(auth service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/register", h.register)
	mux.HandleFunc("POST /api/v1/auth/ekyc/ocr", h.ekycOCR)
	mux.HandleFunc("POST /api/v1/auth/ekyc/confirm", h.ekycConfirm)
	mux.HandleFunc("POST /api/v1/auth/login", h.login)
	mux.HandleFunc("POST /api/v1/auth/logout", h.logout)
}

func (h *AuthHandler) register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.auth.Register(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"message": "registered successfully"})
}

func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	resp, err := h.auth.Login(req)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) logout(w http.ResponseWriter, r *http.Request) {
	token := extractBearerToken(r)
	if token == "" {
		writeError(w, http.StatusUnauthorized, "missing authorization token")
		return
	}
	if err := h.auth.Logout(token); err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

// ekycOCR รับ multipart/form-data (fields: image file + nationalId) หรือ JSON body (nationalId เท่านั้น)
// ถ้ามีไฟล์รูปและมี GEMINI_API_KEY → Gemini Vision อ่านบัตรจริง
// ถ้าไม่มีทั้งคู่ → seed fallback จาก nationalId
func (h *AuthHandler) ekycOCR(w http.ResponseWriter, r *http.Request) {
	var (
		nationalID []byte
		mimeType   string
		imageBytes []byte
		natIDStr   string
	)

	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			writeError(w, http.StatusBadRequest, "failed to parse multipart form")
			return
		}
		natIDStr = r.FormValue("nationalId")

		file, header, err := r.FormFile("image")
		if err == nil {
			defer file.Close()
			imageBytes = make([]byte, header.Size)
			io.ReadFull(file, imageBytes)
			mimeType = header.Header.Get("Content-Type")
		}
	} else {
		var body struct {
			NationalID string `json:"nationalId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		natIDStr = body.NationalID
	}

	_ = nationalID // suppress unused

	if natIDStr == "" && len(imageBytes) == 0 {
		writeError(w, http.StatusBadRequest, "nationalId or image is required")
		return
	}

	result, err := h.auth.ExtractOCR(imageBytes, mimeType, natIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *AuthHandler) ekycConfirm(w http.ResponseWriter, r *http.Request) {
	var req domain.KYCConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.NationalID == "" || req.LaserCode == "" {
		writeError(w, http.StatusBadRequest, "nationalId and laserCode are required")
		return
	}
	if err := h.auth.ConfirmKYC(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"message":   "KYC verified successfully",
		"kycStatus": "verified",
	})
}

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}
