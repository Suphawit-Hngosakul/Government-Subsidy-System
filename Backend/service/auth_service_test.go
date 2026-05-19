package service_test

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"government-subsidy-system/backend/domain"
	"government-subsidy-system/backend/repository"
	"government-subsidy-system/backend/service"
)

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func newAuthService() service.AuthService {
	return service.NewAuthService(
		repository.NewMemoryCitizenRepository(),
		repository.NewMemoryTokenRepository(),
		"", // no Gemini key in tests → seed fallback
	)
}

func TestRegister_Success(t *testing.T) {
	svc := newAuthService()
	err := svc.Register(domain.RegisterRequest{
		NationalID: "1234567890123",
		Password:   "password123",
		Phone:      "0812345678",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRegister_InvalidNationalID(t *testing.T) {
	svc := newAuthService()
	err := svc.Register(domain.RegisterRequest{
		NationalID: "123",
		Password:   "password123",
		Phone:      "0812345678",
	})
	if err == nil {
		t.Fatal("expected error for short national ID")
	}
}

func TestRegister_Duplicate(t *testing.T) {
	svc := newAuthService()
	req := domain.RegisterRequest{
		NationalID: "1234567890123",
		Password:   "password123",
		Phone:      "0812345678",
	}
	_ = svc.Register(req)
	err := svc.Register(req)
	if err == nil {
		t.Fatal("expected error for duplicate registration")
	}
}

func TestRegister_MissingPhone(t *testing.T) {
	svc := newAuthService()
	err := svc.Register(domain.RegisterRequest{
		NationalID: "1234567890123",
		Password:   "password123",
	})
	if err == nil {
		t.Fatal("expected error for missing phone")
	}
}

func TestLogin_Success(t *testing.T) {
	svc := newAuthService()
	_ = svc.Register(domain.RegisterRequest{
		NationalID: "1234567890123",
		Password:   "password123",
		Phone:      "0812345678",
	})
	resp, err := svc.Login(domain.LoginRequest{
		NationalID: "1234567890123",
		Password:   "password123",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if resp.Role != "citizen" {
		t.Fatalf("expected role 'citizen', got '%s'", resp.Role)
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	svc := newAuthService()
	_ = svc.Register(domain.RegisterRequest{
		NationalID: "1234567890123",
		Password:   "password123",
		Phone:      "0812345678",
	})
	_, err := svc.Login(domain.LoginRequest{
		NationalID: "1234567890123",
		Password:   "wrongpassword",
	})
	if err == nil {
		t.Fatal("expected error for invalid password")
	}
}

func TestLogin_NotRegistered(t *testing.T) {
	svc := newAuthService()
	_, err := svc.Login(domain.LoginRequest{
		NationalID: "9999999999999",
		Password:   "password123",
	})
	if err == nil {
		t.Fatal("expected error for unknown citizen")
	}
}

func TestLogout_Success(t *testing.T) {
	svc := newAuthService()
	_ = svc.Register(domain.RegisterRequest{
		NationalID: "1234567890123",
		Password:   "password123",
		Phone:      "0812345678",
	})
	resp, _ := svc.Login(domain.LoginRequest{
		NationalID: "1234567890123",
		Password:   "password123",
	})
	err := svc.Logout(resp.Token)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestLogout_AlreadyRevoked(t *testing.T) {
	svc := newAuthService()
	_ = svc.Register(domain.RegisterRequest{
		NationalID: "1234567890123",
		Password:   "password123",
		Phone:      "0812345678",
	})
	resp, _ := svc.Login(domain.LoginRequest{
		NationalID: "1234567890123",
		Password:   "password123",
	})
	_ = svc.Logout(resp.Token)
	err := svc.Logout(resp.Token)
	if err == nil {
		t.Fatal("expected error for already revoked token")
	}
}

func TestOCRMock_Success(t *testing.T) {
	svc := newAuthService()
	// no image, no key → seed fallback
	result, err := svc.ExtractOCR(nil, "", "1234567890123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.NationalID != "1234567890123" {
		t.Fatalf("expected nationalId to match, got %s", result.NationalID)
	}
	if result.LaserCode == "" {
		t.Fatal("expected non-empty laser code")
	}
}

func TestOCRMock_InvalidNationalID(t *testing.T) {
	svc := newAuthService()
	_, err := svc.ExtractOCR(nil, "", "123")
	if err == nil {
		t.Fatal("expected error for invalid national ID")
	}
}

func TestConfirmKYC_Success(t *testing.T) {
	svc := newAuthService()
	_ = svc.Register(domain.RegisterRequest{
		NationalID: "1234567890123",
		Password:   "password123",
		Phone:      "0812345678",
	})
	ocr, _ := svc.ExtractOCR(nil, "", "1234567890123")
	err := svc.ConfirmKYC(domain.KYCConfirmRequest{
		NationalID:  "1234567890123",
		LaserCode:   ocr.LaserCode,
		FullName:    ocr.FullName,
		DateOfBirth: ocr.DateOfBirth,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestConfirmKYC_NotRegistered(t *testing.T) {
	svc := newAuthService()
	err := svc.ConfirmKYC(domain.KYCConfirmRequest{
		NationalID:  "9999999999999",
		LaserCode:   "AA1-1234567-89",
		FullName:    "test",
		DateOfBirth: "01/01/2533",
	})
	if err == nil {
		t.Fatal("expected error for unknown citizen")
	}
}

func TestValidateJWT_SuccessAndFail(t *testing.T) {
	svc := newAuthService()
	_ = svc.Register(domain.RegisterRequest{
		NationalID: "1234567890123",
		Password:   "password123",
		Phone:      "0812345678",
	})
	resp, err := svc.Login(domain.LoginRequest{
		NationalID: "1234567890123",
		Password:   "password123",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// Validate valid token
	citizenID, role, err := service.ValidateJWT(resp.Token)
	if err != nil {
		t.Fatalf("failed to validate JWT: %v", err)
	}
	if role != "citizen" {
		t.Errorf("expected role 'citizen', got '%s'", role)
	}
	if citizenID == "" {
		t.Error("expected citizenID to be non-empty")
	}

	// Validate invalid token
	_, _, err = service.ValidateJWT("invalid-token-format")
	if err == nil {
		t.Error("expected error for invalid token format, got nil")
	}

	// Validate token with wrong signature
	parts := strings.Split(resp.Token, ".")
	if len(parts) == 3 {
		fakeToken := parts[0] + "." + parts[1] + ".fakesig"
		_, _, err = service.ValidateJWT(fakeToken)
		if err == nil {
			t.Error("expected error for invalid signature, got nil")
		}
	}
}

func TestExtractOCR_GeminiSuccess(t *testing.T) {
	oldTransport := http.DefaultClient.Transport
	defer func() { http.DefaultClient.Transport = oldTransport }()

	mockResponseJSON := `{
		"candidates": [{
			"content": {
				"parts": [{
					"text": "{\n  \"nationalId\": \"1234567890123\",\n  \"fullName\": \"นายสมชาย เข็มทอง\",\n  \"dateOfBirth\": \"01/01/2500\",\n  \"laserCode\": \"JT9-9999999-99\",\n  \"address\": \"123 หมู่ 1\"\n}"
				}]
			}
		}]
	}`

	http.DefaultClient.Transport = roundTripFunc(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(mockResponseJSON)),
			Header:     make(http.Header),
		}
	})


	svc := service.NewAuthService(
		repository.NewMemoryCitizenRepository(),
		repository.NewMemoryTokenRepository(),
		"mock-gemini-key",
	)

	result, err := svc.ExtractOCR([]byte("fake-image-bytes"), "image/jpeg", "1234567890123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.FullName != "นายสมชาย เข็มทอง" || result.NationalID != "1234567890123" {
		t.Errorf("unexpected ocr result: %+v", result)
	}
}

func TestExtractOCR_GeminiErrorFallback(t *testing.T) {
	oldTransport := http.DefaultClient.Transport
	defer func() { http.DefaultClient.Transport = oldTransport }()

	http.DefaultClient.Transport = roundTripFunc(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(bytes.NewBufferString(`{"error": {"message": "invalid key"}}`)),
			Header:     make(http.Header),
		}
	})

	svc := service.NewAuthService(
		repository.NewMemoryCitizenRepository(),
		repository.NewMemoryTokenRepository(),
		"invalid-gemini-key",
	)

	result, err := svc.ExtractOCR([]byte("fake-image-bytes"), "image/jpeg", "1234567890123")
	if err != nil {
		t.Fatalf("expected fallback to work, got error: %v", err)
	}

	if result.FullName != "นาย ทดสอบ ระบบ" {
		t.Errorf("expected fallback to seed ocr, got: %+v", result)
	}
}

