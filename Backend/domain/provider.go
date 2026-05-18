package domain

import "time"

type DOPAVerification struct {
	NationalID   string `json:"nationalId"`
	Valid        bool   `json:"valid"`
	Age          int    `json:"age"`
	Alive        bool   `json:"alive"`
	PersonStatus string `json:"personStatus"`
	CardActive   bool   `json:"cardActive"`
}

type DOPACardStatus struct {
	NationalID    string    `json:"nationalId"`
	CardStatus    string    `json:"cardStatus"`
	CardActive    bool      `json:"cardActive"`
	IssuedAt      time.Time `json:"issuedAt"`
	ExpiredAt     time.Time `json:"expiredAt"`
	RevokedReason *string   `json:"revokedReason,omitempty"`
	CheckedAt     time.Time `json:"checkedAt"`
}

type SSOStatus struct {
	NationalID    string     `json:"nationalId"`
	InsuredStatus string     `json:"insuredStatus"`
	Section       *string    `json:"section,omitempty"`
	Insured       bool       `json:"insured"`
	EmployerID    *string    `json:"employerId,omitempty"`
	EmployerName  *string    `json:"employerName,omitempty"`
	RegisteredAt  *time.Time `json:"registeredAt,omitempty"`
}

type SSOContribution struct {
	NationalID              string             `json:"nationalId"`
	ContributionMonths      int                `json:"contributionMonths"`
	LatestContributionMonth *time.Time         `json:"latestContributionMonth,omitempty"`
	RecentContributions     []ContributionItem `json:"recentContributions"`
}

type ContributionItem struct {
	ContributionMonth time.Time  `json:"contributionMonth"`
	EmployeeAmount    float64    `json:"employeeAmount"`
	EmployerAmount    float64    `json:"employerAmount"`
	GovernmentAmount  float64    `json:"governmentAmount"`
	PaidAt            *time.Time `json:"paidAt,omitempty"`
	PaymentStatus     string     `json:"paymentStatus"`
}

type KTBFinancialCheck struct {
	NationalID           string  `json:"nationalId"`
	DepositTotal         float64 `json:"depositTotal"`
	AverageMonthlyIncome float64 `json:"averageMonthlyIncome"`
	ActiveAccountCount   int     `json:"activeAccountCount"`
}

type KTBAccountStatus struct {
	NationalID       string            `json:"nationalId"`
	HasActiveAccount bool              `json:"hasActiveAccount"`
	PromptPayLinked  bool              `json:"promptPayLinked"`
	Accounts         []BankAccountItem `json:"accounts"`
	PromptPay        []PromptPayItem   `json:"promptPay"`
}

type BankAccountItem struct {
	BankCode      string  `json:"bankCode"`
	BranchCode    *string `json:"branchCode,omitempty"`
	AccountNo     string  `json:"accountNo"`
	AccountName   string  `json:"accountName"`
	AccountType   string  `json:"accountType"`
	AccountStatus string  `json:"accountStatus"`
	Balance       float64 `json:"balance"`
}

type PromptPayItem struct {
	ProxyType          string     `json:"proxyType"`
	ProxyValue         string     `json:"proxyValue"`
	RegistrationStatus string     `json:"registrationStatus"`
	RegisteredAt       time.Time  `json:"registeredAt"`
	RevokedAt          *time.Time `json:"revokedAt,omitempty"`
}
