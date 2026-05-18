package service

import (
	"context"
	"testing"

	"government-subsidy-system/backend/adapter"
	"government-subsidy-system/backend/domain"
	"government-subsidy-system/backend/repository"
)

func TestEvaluateDecisionApprovesEligibleClaim(t *testing.T) {
	result := EvaluateDecision("claim-1", domain.EligibilitySources{
		DOPA: domain.DOPAResult{Valid: true, Age: 35, Alive: true, CardActive: true},
		SSO:  domain.SSOResult{Section: "40", ContributionMonths: 12},
		KTB:  domain.KTBResult{DepositTotal: 12000, AverageMonthlyIncome: 15000, PromptPayLinked: true},
	}, nil)

	if result.Status != domain.StatusApproved {
		t.Fatalf("expected %s, got %s", domain.StatusApproved, result.Status)
	}
}

func TestEvaluateDecisionRejectsInvalidDOPA(t *testing.T) {
	result := EvaluateDecision("claim-1", domain.EligibilitySources{
		DOPA: domain.DOPAResult{Valid: false, Age: 35, Alive: true, CardActive: true},
		SSO:  domain.SSOResult{Section: "40", ContributionMonths: 12},
		KTB:  domain.KTBResult{DepositTotal: 12000, AverageMonthlyIncome: 15000, PromptPayLinked: true},
	}, nil)

	if result.Status != domain.StatusRejected {
		t.Fatalf("expected %s, got %s", domain.StatusRejected, result.Status)
	}
}

func TestOrchestrateStoresResultAndEvents(t *testing.T) {
	repo := repository.NewMemoryClaimRepository()
	svc := NewOrchestratorService(
		repo,
		adapter.NewMockDOPAAdapter(),
		adapter.NewMockSSOAdapter(),
		adapter.NewMockKTBAdapter(),
	)

	result, err := svc.Orchestrate(context.Background(), domain.OrchestrateRequest{
		ClaimID:    "claim-1",
		NationalID: "1101700203451",
	})
	if err != nil {
		t.Fatalf("orchestrate returned error: %v", err)
	}
	if result.Status != domain.StatusApproved {
		t.Fatalf("expected %s, got %s", domain.StatusApproved, result.Status)
	}

	stored, ok := repo.GetResult("claim-1")
	if !ok {
		t.Fatal("expected result to be stored")
	}
	if stored.ClaimID != "claim-1" {
		t.Fatalf("expected stored claim id claim-1, got %s", stored.ClaimID)
	}

	events := repo.Events("claim-1")
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
}
