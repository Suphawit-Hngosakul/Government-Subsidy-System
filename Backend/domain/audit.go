package domain

import "time"

type AuditAction string

const (
	AuditActionClaimApproved AuditAction = "claim.approved"
	AuditActionClaimRejected AuditAction = "claim.rejected"
)

type AuditEntry struct {
	ID       string         `json:"id"`
	At       time.Time      `json:"at"`
	Actor    string         `json:"actor"`
	Action   AuditAction    `json:"action"`
	EntityID string         `json:"entityId"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type Stats struct {
	Projects ProjectStats `json:"projects"`
	Claims   ClaimStats   `json:"claims"`
}

type ProjectStats struct {
	Total  int `json:"total"`
	Active int `json:"active"`
}

type ClaimStats struct {
	Total    int `json:"total"`
	Pending  int `json:"pending"`
	Approved int `json:"approved"`
	Rejected int `json:"rejected"`
}
