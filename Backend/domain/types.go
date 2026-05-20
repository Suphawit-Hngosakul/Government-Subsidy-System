package domain

import "time"

type ClaimStatus string

const (
	StatusProcessing ClaimStatus = "processing"
	StatusApproved   ClaimStatus = "approved"
	StatusRejected   ClaimStatus = "rejected"
	StatusPending    ClaimStatus = "pending"
)

// ============= Benefit / Citizen Claim Models =============

// Claim represents a citizen's benefit claim (citizen side view)
type Claim struct {
	ID          string             `json:"id"`
	NationalID  string             `json:"nationalId"`
	ProjectID   string             `json:"projectId"`
	Status      ClaimStatus        `json:"status"`
	SubmittedAt time.Time          `json:"submittedAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
	Eligibility *EligibilityResult `json:"eligibility,omitempty"`
	Project     *Project           `json:"project,omitempty"`
}

type BenefitClaimRequest struct {
	NationalID string `json:"nationalId"`
	ProjectID  string `json:"projectId"`
}

type ClaimResponse struct {
	ID          string             `json:"id"`
	NationalID  string             `json:"nationalId"`
	ProjectID   string             `json:"projectId"`
	Status      ClaimStatus        `json:"status"`
	SubmittedAt time.Time          `json:"submittedAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
	Eligibility *EligibilityResult `json:"eligibility,omitempty"`
	Project     *Project           `json:"project,omitempty"`
}

// ============= Orchestrator Models =============

type OrchestrateRequest struct {
	ClaimID    string `json:"claimId"`
	NationalID string `json:"nationalId"`
	ProjectID  string `json:"projectId,omitempty"`
}

type DecisionRequest struct {
	ClaimID string `json:"claimId"`
}

// DOPAResult, SSOResult, KTBResult are aliases to the Eligibility* types in claim.go
type DOPAResult = EligibilityDOPA
type SSOResult = EligibilitySSO
type KTBResult = EligibilityKTB

type DecisionResult struct {
	ClaimID string             `json:"claimId"`
	Status  ClaimStatus        `json:"status"`
	Reasons []string           `json:"reasons"`
	Sources EligibilitySources `json:"sources"`
}

type StatusEvent struct {
	ClaimID string      `json:"claimId"`
	Status  ClaimStatus `json:"status"`
	Message string      `json:"message"`
	At      time.Time   `json:"at"`
}
