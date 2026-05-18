package service

import (
	"context"
	"errors"
	"regexp"

	"government-subsidy-system/backend/domain"
	"government-subsidy-system/backend/repository"
)

var (
	ErrInvalidNationalID  = errors.New("nationalId must be 13 digits")
	ErrNationalIDRequired = errors.New("nationalId is required")
)

var nationalIDPattern = regexp.MustCompile(`^[0-9]{13}$`)

type ProviderService struct {
	repo repository.ProviderRepository
}

func NewProviderService(repo repository.ProviderRepository) *ProviderService {
	return &ProviderService{repo: repo}
}

func (s *ProviderService) VerifyDOPA(ctx context.Context, nationalID string) (domain.DOPAVerification, error) {
	if err := validateNationalID(nationalID); err != nil {
		return domain.DOPAVerification{}, err
	}
	return s.repo.GetDOPAVerification(ctx, nationalID)
}

func (s *ProviderService) DOPACardStatus(ctx context.Context, nationalID string) (domain.DOPACardStatus, error) {
	if err := validateNationalID(nationalID); err != nil {
		return domain.DOPACardStatus{}, err
	}
	return s.repo.GetDOPACardStatus(ctx, nationalID)
}

func (s *ProviderService) SSOStatus(ctx context.Context, nationalID string) (domain.SSOStatus, error) {
	if err := validateNationalID(nationalID); err != nil {
		return domain.SSOStatus{}, err
	}
	return s.repo.GetSSOStatus(ctx, nationalID)
}

func (s *ProviderService) SSOContribution(ctx context.Context, nationalID string) (domain.SSOContribution, error) {
	if err := validateNationalID(nationalID); err != nil {
		return domain.SSOContribution{}, err
	}
	return s.repo.GetSSOContribution(ctx, nationalID)
}

func (s *ProviderService) KTBFinancialCheck(ctx context.Context, nationalID string) (domain.KTBFinancialCheck, error) {
	if err := validateNationalID(nationalID); err != nil {
		return domain.KTBFinancialCheck{}, err
	}
	return s.repo.GetKTBFinancialCheck(ctx, nationalID)
}

func (s *ProviderService) KTBAccountStatus(ctx context.Context, nationalID string) (domain.KTBAccountStatus, error) {
	if err := validateNationalID(nationalID); err != nil {
		return domain.KTBAccountStatus{}, err
	}
	return s.repo.GetKTBAccountStatus(ctx, nationalID)
}

func (s *ProviderService) Verify(ctx context.Context, nationalID string) (domain.DOPAResult, error) {
	result, err := s.VerifyDOPA(ctx, nationalID)
	if err != nil {
		return domain.DOPAResult{}, err
	}
	return domain.DOPAResult{
		Valid:      result.Valid,
		Age:        result.Age,
		Alive:      result.Alive,
		CardActive: result.CardActive,
	}, nil
}

func (s *ProviderService) Status(ctx context.Context, nationalID string) (domain.SSOResult, error) {
	status, err := s.SSOStatus(ctx, nationalID)
	if err != nil {
		return domain.SSOResult{}, err
	}
	contribution, err := s.SSOContribution(ctx, nationalID)
	if err != nil {
		return domain.SSOResult{}, err
	}

	section := ""
	if status.Section != nil {
		section = *status.Section
	}
	return domain.SSOResult{
		Section:            section,
		ContributionMonths: contribution.ContributionMonths,
	}, nil
}

func (s *ProviderService) FinancialCheck(ctx context.Context, nationalID string) (domain.KTBResult, error) {
	financial, err := s.KTBFinancialCheck(ctx, nationalID)
	if err != nil {
		return domain.KTBResult{}, err
	}
	account, err := s.KTBAccountStatus(ctx, nationalID)
	if err != nil {
		return domain.KTBResult{}, err
	}
	return domain.KTBResult{
		DepositTotal:         financial.DepositTotal,
		AverageMonthlyIncome: financial.AverageMonthlyIncome,
		PromptPayLinked:      account.PromptPayLinked,
	}, nil
}

func validateNationalID(nationalID string) error {
	if nationalID == "" {
		return ErrNationalIDRequired
	}
	if !nationalIDPattern.MatchString(nationalID) {
		return ErrInvalidNationalID
	}
	return nil
}
