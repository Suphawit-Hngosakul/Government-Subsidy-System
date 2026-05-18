package service_test

import (
	"testing"

	"government-subsidy-system/backend/domain"
	"government-subsidy-system/backend/repository"
	"government-subsidy-system/backend/service"
)

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
