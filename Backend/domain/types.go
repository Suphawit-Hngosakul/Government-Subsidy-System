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

type DOPAResult struct {
	Valid      bool `json:"valid"`
	Age        int  `json:"age"`
	Alive      bool `json:"alive"`
	CardActive bool `json:"cardActive"`
}

type SSOResult struct {
	Section            string `json:"section"`
	ContributionMonths int    `json:"contributionMonths"`
}

type KTBResult struct {
	DepositTotal         float64 `json:"depositTotal"`
	AverageMonthlyIncome float64 `json:"averageMonthlyIncome"`
	PromptPayLinked      bool    `json:"promptPayLinked"`
}

type EligibilitySources struct {
	DOPA DOPAResult `json:"dopa"`
	SSO  SSOResult  `json:"sso"`
	KTB  KTBResult  `json:"ktb"`
}

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
