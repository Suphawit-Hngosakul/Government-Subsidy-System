package domain

import "time"

type ClaimStatus string

const (
	StatusProcessing ClaimStatus = "processing"
	StatusApproved   ClaimStatus = "approved"
	StatusRejected   ClaimStatus = "rejected"
	StatusPending    ClaimStatus = "pending"
)

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
