package domain

import "time"

type KYCStatus string

const (
	KYCStatusPending  KYCStatus = "pending"
	KYCStatusVerified KYCStatus = "verified"
	KYCStatusRejected KYCStatus = "rejected"
)

type Citizen struct {
	ID         string    `json:"id"`
	NationalID string    `json:"nationalId"`
	LaserCode  string    `json:"-"`
	Phone      string    `json:"phone"`
	HashedPin  string    `json:"-"`
	KYCStatus  KYCStatus `json:"kycStatus"`
	CreatedAt  time.Time `json:"createdAt"`
}

type RegisterRequest struct {
	NationalID string `json:"nationalId"`
	Password   string `json:"password"`
	Phone      string `json:"phone"`
}

type LoginRequest struct {
	NationalID string `json:"nationalId"`
	Password   string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	Role  string `json:"role"`
}

type OCRResult struct {
	NationalID  string `json:"nationalId"`
	FullName    string `json:"fullName"`
	DateOfBirth string `json:"dateOfBirth"`
	LaserCode   string `json:"laserCode"`
	Address     string `json:"address"`
}

type KYCConfirmRequest struct {
	NationalID  string `json:"nationalId"`
	LaserCode   string `json:"laserCode"`
	FullName    string `json:"fullName"`
	DateOfBirth string `json:"dateOfBirth"`
}
