package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"government-subsidy-system/backend/domain"
	"government-subsidy-system/backend/repository"
)

// jwtSecret signs tokens in v1; replace with env var in production.
var jwtSecret = []byte("govt-subsidy-secret-v1")

type AuthService interface {
	Register(req domain.RegisterRequest) error
	Login(req domain.LoginRequest) (*domain.LoginResponse, error)
	Logout(token string) error
	// ExtractOCR อ่านข้อมูลจากรูปบัตรผ่าน Gemini Vision ถ้ามี API key และมีรูป
	// ถ้าไม่มีทั้งคู่จะ fallback เป็น seed mock จาก nationalID
	ExtractOCR(imageBytes []byte, mimeType string, nationalID string) (*domain.OCRResult, error)
	ConfirmKYC(req domain.KYCConfirmRequest) error
}

type authService struct {
	citizens  repository.CitizenRepository
	tokens    repository.TokenRepository
	geminiKey string
}

func NewAuthService(citizens repository.CitizenRepository, tokens repository.TokenRepository, geminiKey string) AuthService {
	return &authService{citizens: citizens, tokens: tokens, geminiKey: geminiKey}
}

func (s *authService) Register(req domain.RegisterRequest) error {
	if len(req.NationalID) != 13 {
		return fmt.Errorf("national ID must be 13 digits")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	if req.Phone == "" {
		return fmt.Errorf("phone is required")
	}
	citizen := &domain.Citizen{
		NationalID: req.NationalID,
		Phone:      req.Phone,
		HashedPin:  hashPassword(req.Password),
		KYCStatus:  domain.KYCStatusPending,
	}
	return s.citizens.Create(citizen)
}

func (s *authService) Login(req domain.LoginRequest) (*domain.LoginResponse, error) {
	citizen, err := s.citizens.FindByNationalID(req.NationalID)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	if hashPassword(req.Password) != citizen.HashedPin {
		return nil, fmt.Errorf("invalid credentials")
	}
	token, err := generateJWT(citizen.ID, "citizen")
	if err != nil {
		return nil, err
	}
	s.tokens.Save(token, citizen.ID)
	return &domain.LoginResponse{Token: token, Role: "citizen"}, nil
}

func (s *authService) Logout(token string) error {
	_, ok := s.tokens.FindCitizenID(token)
	if !ok {
		return fmt.Errorf("token not found or already revoked")
	}
	s.tokens.Revoke(token)
	return nil
}

func (s *authService) ExtractOCR(imageBytes []byte, mimeType string, nationalID string) (*domain.OCRResult, error) {
	// ถ้ามีรูปและมี Gemini key → ส่งไปอ่านจริง
	if len(imageBytes) > 0 && s.geminiKey != "" {
		result, err := callGeminiVision(s.geminiKey, imageBytes, mimeType)
		if err == nil {
			return result, nil
		}
		log.Printf("[OCR] Gemini error: %v — using seed fallback", err)
	}

	// seed fallback
	if len(nationalID) != 13 {
		return nil, fmt.Errorf("national ID must be 13 digits")
	}
	return &domain.OCRResult{
		NationalID:  nationalID,
		FullName:    "นาย ทดสอบ ระบบ",
		DateOfBirth: "01/01/2533",
		LaserCode:   "AA1-1234567-89",
		Address:     "1/1 ถ.สุขุมวิท แขวงคลองเตย เขตคลองเตย กรุงเทพมหานคร 10110",
	}, nil
}

func (s *authService) ConfirmKYC(req domain.KYCConfirmRequest) error {
	if _, err := s.citizens.FindByNationalID(req.NationalID); err != nil {
		return err
	}
	if err := s.citizens.UpdateLaserCode(req.NationalID, req.LaserCode); err != nil {
		return err
	}
	return s.citizens.UpdateKYCStatus(req.NationalID, domain.KYCStatusVerified)
}

// ValidateJWT parses and verifies a token produced by generateJWT.
func ValidateJWT(token string) (citizenID string, role string, err error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", "", errors.New("invalid token format")
	}
	sigInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, jwtSecret)
	mac.Write([]byte(sigInput))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return "", "", errors.New("invalid token signature")
	}
	pb, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", err
	}
	var payload jwtPayload
	if err := json.Unmarshal(pb, &payload); err != nil {
		return "", "", err
	}
	if time.Now().Unix() > payload.Exp {
		return "", "", errors.New("token expired")
	}
	return payload.Sub, payload.Role, nil
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type jwtPayload struct {
	Sub  string `json:"sub"`
	Role string `json:"role"`
	Exp  int64  `json:"exp"`
	Jti  string `json:"jti"`
}

func generateJWT(citizenID, role string) (string, error) {
	hb, err := json.Marshal(jwtHeader{Alg: "HS256", Typ: "JWT"})
	if err != nil {
		return "", err
	}
	pb, err := json.Marshal(jwtPayload{
		Sub:  citizenID,
		Role: role,
		Exp:  time.Now().Add(24 * time.Hour).Unix(),
		Jti:  randomID(),
	})
	if err != nil {
		return "", err
	}
	headerEnc := base64.RawURLEncoding.EncodeToString(hb)
	payloadEnc := base64.RawURLEncoding.EncodeToString(pb)
	sigInput := headerEnc + "." + payloadEnc
	mac := hmac.New(sha256.New, jwtSecret)
	mac.Write([]byte(sigInput))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return sigInput + "." + sig, nil
}

func hashPassword(password string) string {
	h := sha256.Sum256([]byte(password))
	return base64.StdEncoding.EncodeToString(h[:])
}

func randomID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
