package service

import (
	"context"
	"testing"

	"government-subsidy-system/backend/domain"
	"government-subsidy-system/backend/repository"
)

type stubProviderRepository struct{}

func (stubProviderRepository) GetDOPAVerification(context.Context, string) (domain.DOPAVerification, error) {
	return domain.DOPAVerification{NationalID: "1101700203451", Valid: true, Age: 35, Alive: true, CardActive: true}, nil
}
func (stubProviderRepository) GetDOPACardStatus(context.Context, string) (domain.DOPACardStatus, error) {
	return domain.DOPACardStatus{}, nil
}
func (stubProviderRepository) GetSSOStatus(context.Context, string) (domain.SSOStatus, error) {
	section := "40"
	return domain.SSOStatus{NationalID: "1101700203451", InsuredStatus: "insured", Section: &section, Insured: true}, nil
}
func (stubProviderRepository) GetSSOContribution(context.Context, string) (domain.SSOContribution, error) {
	return domain.SSOContribution{NationalID: "1101700203451", ContributionMonths: 12}, nil
}
func (stubProviderRepository) GetKTBFinancialCheck(context.Context, string) (domain.KTBFinancialCheck, error) {
	return domain.KTBFinancialCheck{NationalID: "1101700203451", DepositTotal: 12000, AverageMonthlyIncome: 15000, ActiveAccountCount: 1}, nil
}
func (stubProviderRepository) GetKTBAccountStatus(context.Context, string) (domain.KTBAccountStatus, error) {
	return domain.KTBAccountStatus{NationalID: "1101700203451", HasActiveAccount: true, PromptPayLinked: true}, nil
}

var _ repository.ProviderRepository = stubProviderRepository{}

func TestProviderServiceRejectsInvalidNationalID(t *testing.T) {
	svc := NewProviderService(stubProviderRepository{})
	_, err := svc.VerifyDOPA(context.Background(), "123")
	if err != ErrInvalidNationalID {
		t.Fatalf("expected ErrInvalidNationalID, got %v", err)
	}
}

func TestProviderServiceImplementsOrchestratorClients(t *testing.T) {
	svc := NewProviderService(stubProviderRepository{})

	dopa, err := svc.Verify(context.Background(), "1101700203451")
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}
	if !dopa.Valid || !dopa.CardActive {
		t.Fatalf("unexpected dopa result: %+v", dopa)
	}

	sso, err := svc.Status(context.Background(), "1101700203451")
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}
	if sso.Section != "40" || sso.ContributionMonths != 12 {
		t.Fatalf("unexpected sso result: %+v", sso)
	}

	ktb, err := svc.FinancialCheck(context.Background(), "1101700203451")
	if err != nil {
		t.Fatalf("FinancialCheck returned error: %v", err)
	}
	if !ktb.PromptPayLinked || ktb.DepositTotal != 12000 {
		t.Fatalf("unexpected ktb result: %+v", ktb)
	}
}
