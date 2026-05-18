package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"government-subsidy-system/backend/domain"
)

var ErrProviderRecordNotFound = errors.New("provider record not found")

type ProviderRepository interface {
	GetDOPAVerification(ctx context.Context, nationalID string) (domain.DOPAVerification, error)
	GetDOPACardStatus(ctx context.Context, nationalID string) (domain.DOPACardStatus, error)
	GetSSOStatus(ctx context.Context, nationalID string) (domain.SSOStatus, error)
	GetSSOContribution(ctx context.Context, nationalID string) (domain.SSOContribution, error)
	GetKTBFinancialCheck(ctx context.Context, nationalID string) (domain.KTBFinancialCheck, error)
	GetKTBAccountStatus(ctx context.Context, nationalID string) (domain.KTBAccountStatus, error)
}

type PostgresProviderRepository struct {
	db *sql.DB
}

func NewPostgresProviderRepository(db *sql.DB) *PostgresProviderRepository {
	return &PostgresProviderRepository{db: db}
}

func (r *PostgresProviderRepository) GetDOPAVerification(ctx context.Context, nationalID string) (domain.DOPAVerification, error) {
	const query = `
SELECT
	c.national_id,
	EXTRACT(YEAR FROM age(CURRENT_DATE, c.date_of_birth))::int AS age,
	c.person_status,
	COALESCE(card.card_status, '') AS card_status
FROM dopa.citizens c
LEFT JOIN LATERAL (
	SELECT card_status
	FROM dopa.id_cards
	WHERE national_id = c.national_id
	ORDER BY issued_at DESC, card_id DESC
	LIMIT 1
) card ON true
WHERE c.national_id = $1`

	var result domain.DOPAVerification
	var cardStatus string
	if err := r.db.QueryRowContext(ctx, query, nationalID).Scan(
		&result.NationalID,
		&result.Age,
		&result.PersonStatus,
		&cardStatus,
	); err != nil {
		return domain.DOPAVerification{}, mapNotFound(err)
	}

	result.Alive = result.PersonStatus == "alive"
	result.CardActive = cardStatus == "active"
	result.Valid = result.Alive && result.CardActive
	return result, nil
}

func (r *PostgresProviderRepository) GetDOPACardStatus(ctx context.Context, nationalID string) (domain.DOPACardStatus, error) {
	const query = `
SELECT national_id, issued_at, expired_at, card_status, revoked_reason
FROM dopa.id_cards
WHERE national_id = $1
ORDER BY issued_at DESC, card_id DESC
LIMIT 1`

	var result domain.DOPACardStatus
	var revoked sql.NullString
	if err := r.db.QueryRowContext(ctx, query, nationalID).Scan(
		&result.NationalID,
		&result.IssuedAt,
		&result.ExpiredAt,
		&result.CardStatus,
		&revoked,
	); err != nil {
		return domain.DOPACardStatus{}, mapNotFound(err)
	}

	result.CardActive = result.CardStatus == "active" && result.ExpiredAt.After(time.Now())
	result.CheckedAt = time.Now().UTC()
	result.RevokedReason = stringPtrFromNull(revoked)
	return result, nil
}

func (r *PostgresProviderRepository) GetSSOStatus(ctx context.Context, nationalID string) (domain.SSOStatus, error) {
	const query = `
SELECT national_id, insured_status, section, employer_id, employer_name, registered_at
FROM sso.insured_persons
WHERE national_id = $1`

	var result domain.SSOStatus
	var section, employerID, employerName sql.NullString
	var registeredAt sql.NullTime
	if err := r.db.QueryRowContext(ctx, query, nationalID).Scan(
		&result.NationalID,
		&result.InsuredStatus,
		&section,
		&employerID,
		&employerName,
		&registeredAt,
	); err != nil {
		return domain.SSOStatus{}, mapNotFound(err)
	}

	result.Section = stringPtrFromNull(section)
	result.EmployerID = stringPtrFromNull(employerID)
	result.EmployerName = stringPtrFromNull(employerName)
	result.RegisteredAt = timePtrFromNull(registeredAt)
	result.Insured = result.InsuredStatus == "insured"
	return result, nil
}

func (r *PostgresProviderRepository) GetSSOContribution(ctx context.Context, nationalID string) (domain.SSOContribution, error) {
	const summaryQuery = `
SELECT national_id, contribution_months, latest_contribution_month
FROM sso.insured_persons
WHERE national_id = $1`

	var result domain.SSOContribution
	var latest sql.NullTime
	if err := r.db.QueryRowContext(ctx, summaryQuery, nationalID).Scan(
		&result.NationalID,
		&result.ContributionMonths,
		&latest,
	); err != nil {
		return domain.SSOContribution{}, mapNotFound(err)
	}
	result.LatestContributionMonth = timePtrFromNull(latest)

	const listQuery = `
SELECT contribution_month, employee_amount, employer_amount, government_amount, paid_at, payment_status
FROM sso.contributions
WHERE national_id = $1
ORDER BY contribution_month DESC
LIMIT 12`

	rows, err := r.db.QueryContext(ctx, listQuery, nationalID)
	if err != nil {
		return domain.SSOContribution{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.ContributionItem
		var paidAt sql.NullTime
		if err := rows.Scan(
			&item.ContributionMonth,
			&item.EmployeeAmount,
			&item.EmployerAmount,
			&item.GovernmentAmount,
			&paidAt,
			&item.PaymentStatus,
		); err != nil {
			return domain.SSOContribution{}, err
		}
		item.PaidAt = timePtrFromNull(paidAt)
		result.RecentContributions = append(result.RecentContributions, item)
	}
	if err := rows.Err(); err != nil {
		return domain.SSOContribution{}, err
	}
	return result, nil
}

func (r *PostgresProviderRepository) GetKTBFinancialCheck(ctx context.Context, nationalID string) (domain.KTBFinancialCheck, error) {
	const query = `
SELECT
	national_id,
	COALESCE(SUM(balance), 0) AS deposit_total,
	COALESCE(MAX(average_monthly_income), 0) AS average_monthly_income,
	COUNT(*) FILTER (WHERE account_status = 'active') AS active_account_count
FROM ktb.bank_accounts
WHERE national_id = $1
GROUP BY national_id`

	var result domain.KTBFinancialCheck
	if err := r.db.QueryRowContext(ctx, query, nationalID).Scan(
		&result.NationalID,
		&result.DepositTotal,
		&result.AverageMonthlyIncome,
		&result.ActiveAccountCount,
	); err != nil {
		return domain.KTBFinancialCheck{}, mapNotFound(err)
	}
	return result, nil
}

func (r *PostgresProviderRepository) GetKTBAccountStatus(ctx context.Context, nationalID string) (domain.KTBAccountStatus, error) {
	accounts, err := r.ktbAccounts(ctx, nationalID)
	if err != nil {
		return domain.KTBAccountStatus{}, err
	}
	if len(accounts) == 0 {
		return domain.KTBAccountStatus{}, ErrProviderRecordNotFound
	}

	promptPay, err := r.promptPayRegistrations(ctx, nationalID)
	if err != nil {
		return domain.KTBAccountStatus{}, err
	}

	result := domain.KTBAccountStatus{
		NationalID: nationalID,
		Accounts:   accounts,
		PromptPay:  promptPay,
	}
	for _, account := range accounts {
		if account.AccountStatus == "active" {
			result.HasActiveAccount = true
			break
		}
	}
	for _, item := range promptPay {
		if item.RegistrationStatus == "active" {
			result.PromptPayLinked = true
			break
		}
	}
	return result, nil
}

func (r *PostgresProviderRepository) ktbAccounts(ctx context.Context, nationalID string) ([]domain.BankAccountItem, error) {
	const query = `
SELECT bank_code, branch_code, account_no, account_name, account_type, account_status, balance
FROM ktb.bank_accounts
WHERE national_id = $1
ORDER BY account_id`

	rows, err := r.db.QueryContext(ctx, query, nationalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []domain.BankAccountItem
	for rows.Next() {
		var item domain.BankAccountItem
		var branch sql.NullString
		if err := rows.Scan(
			&item.BankCode,
			&branch,
			&item.AccountNo,
			&item.AccountName,
			&item.AccountType,
			&item.AccountStatus,
			&item.Balance,
		); err != nil {
			return nil, err
		}
		item.BranchCode = stringPtrFromNull(branch)
		accounts = append(accounts, item)
	}
	return accounts, rows.Err()
}

func (r *PostgresProviderRepository) promptPayRegistrations(ctx context.Context, nationalID string) ([]domain.PromptPayItem, error) {
	const query = `
SELECT proxy_type, proxy_value, registration_status, registered_at, revoked_at
FROM ktb.promptpay_registrations
WHERE national_id = $1
ORDER BY promptpay_id`

	rows, err := r.db.QueryContext(ctx, query, nationalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var registrations []domain.PromptPayItem
	for rows.Next() {
		var item domain.PromptPayItem
		var revokedAt sql.NullTime
		if err := rows.Scan(
			&item.ProxyType,
			&item.ProxyValue,
			&item.RegistrationStatus,
			&item.RegisteredAt,
			&revokedAt,
		); err != nil {
			return nil, err
		}
		item.RevokedAt = timePtrFromNull(revokedAt)
		registrations = append(registrations, item)
	}
	return registrations, rows.Err()
}

func mapNotFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrProviderRecordNotFound
	}
	return err
}

func stringPtrFromNull(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func timePtrFromNull(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}
