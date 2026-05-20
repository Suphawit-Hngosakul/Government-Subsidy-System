package domain

import "time"

type Project struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Active      bool            `json:"active"`
	Criteria    ProjectCriteria `json:"criteria"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

type ProjectCriteria struct {
	MinAge             int      `json:"minAge"`
	MaxAge             int      `json:"maxAge,omitempty"`
	MaxMonthlyIncome   float64  `json:"maxMonthlyIncome,omitempty"`
	MaxDepositTotal    float64  `json:"maxDepositTotal,omitempty"`
	AllowedSSOSections []string `json:"allowedSsoSections,omitempty"`
	RequirePromptPay   bool     `json:"requirePromptPay"`
	RequireSSO         bool     `json:"requireSso"`
	RequireKTB         bool     `json:"requireKtb"`
}

type ProjectInput struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Active      bool            `json:"active"`
	Criteria    ProjectCriteria `json:"criteria"`
}

type ProjectUpdate struct {
	Name        *string          `json:"name,omitempty"`
	Description *string          `json:"description,omitempty"`
	Active      *bool            `json:"active,omitempty"`
	Criteria    *ProjectCriteria `json:"criteria,omitempty"`
}
