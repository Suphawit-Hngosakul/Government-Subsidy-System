package domain

import "time"

type OfficerClaimStatus string

const (
	OfficerStatusPending  OfficerClaimStatus = "pending"
	OfficerStatusApproved OfficerClaimStatus = "approved"
	OfficerStatusRejected OfficerClaimStatus = "rejected"
)

type OfficerClaim struct {
	ClaimID         string             `json:"claimId"`
	NationalID      string             `json:"nationalId"`
	ProjectID       string             `json:"projectId"`
	Status          OfficerClaimStatus `json:"status"`
	SubmittedAt     time.Time          `json:"submittedAt"`
	OfficerDecision *OfficerDecision   `json:"officerDecision,omitempty"`
}

type OfficerDecision struct {
	OfficerID string    `json:"officerId"`
	Reason    string    `json:"reason"`
	DecidedAt time.Time `json:"decidedAt"`
}

type OfficerDecisionInput struct {
	OfficerID string `json:"officerId"`
	Reason    string `json:"reason"`
}

type OfficerClaimDetail struct {
	Claim       OfficerClaim       `json:"claim"`
	Eligibility *EligibilityResult `json:"eligibility,omitempty"`
}

type EligibilityResult struct {
	ClaimID string             `json:"claimId"`
	Status  ClaimStatus        `json:"status"`
	Reasons []string           `json:"reasons"`
	Sources EligibilitySources `json:"sources"`
}

type EligibilitySources struct {
	DOPA EligibilityDOPA `json:"dopa"`
	SSO  EligibilitySSO  `json:"sso"`
	KTB  EligibilityKTB  `json:"ktb"`
}

type EligibilityDOPA struct {
	Valid      bool `json:"valid"`
	Age        int  `json:"age"`
	Alive      bool `json:"alive"`
	CardActive bool `json:"cardActive"`
}

type EligibilitySSO struct {
	Section            string `json:"section"`
	ContributionMonths int    `json:"contributionMonths"`
}

type EligibilityKTB struct {
	DepositTotal         float64 `json:"depositTotal"`
	AverageMonthlyIncome float64 `json:"averageMonthlyIncome"`
	PromptPayLinked      bool    `json:"promptPayLinked"`
}
