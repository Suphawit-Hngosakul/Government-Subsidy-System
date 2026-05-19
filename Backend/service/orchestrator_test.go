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

func TestOrchestratorDecisionEventsSubscribe(t *testing.T) {
	repo := repository.NewMemoryClaimRepository()
	svc := NewOrchestratorService(
		repo,
		adapter.NewMockDOPAAdapter(),
		adapter.NewMockSSOAdapter(),
		adapter.NewMockKTBAdapter(),
	)

	// Decision on unknown claim should return error
	_, err := svc.Decision(context.Background(), domain.DecisionRequest{ClaimID: "unknown"})
	if err == nil {
		t.Error("expected error for unknown claim decision, got nil")
	}

	// Decision with empty claim ID
	_, err = svc.Decision(context.Background(), domain.DecisionRequest{ClaimID: ""})
	if err == nil {
		t.Error("expected error for empty claim ID, got nil")
	}

	// Orchestrate to seed the claim
	_, err = svc.Orchestrate(context.Background(), domain.OrchestrateRequest{
		ClaimID:    "claim-123",
		NationalID: "1101700203451",
	})
	if err != nil {
		t.Fatalf("failed to orchestrate: %v", err)
	}

	// Now check Decision
	dec, err := svc.Decision(context.Background(), domain.DecisionRequest{ClaimID: "claim-123"})
	if err != nil {
		t.Errorf("failed to get decision: %v", err)
	}
	if dec.ClaimID != "claim-123" {
		t.Errorf("expected claim-123, got %s", dec.ClaimID)
	}

	// Events
	events := svc.Events("claim-123")
	if len(events) == 0 {
		t.Error("expected status events, got 0")
	}

	// Subscribe
	ch, cancel := svc.Subscribe("claim-123")
	if ch == nil {
		t.Error("expected subscribe channel, got nil")
	}
	cancel()
}

func TestEvaluateDecisionAllRules(t *testing.T) {
	// Rule: provider error
	r1 := EvaluateDecision("claim-1", domain.EligibilitySources{}, []string{"some error"})
	if r1.Status != domain.StatusPending || r1.Reasons[0] != "some error" {
		t.Errorf("expected Pending with reason, got status %s, reasons %v", r1.Status, r1.Reasons)
	}

	// Rule: citizen age under 18
	r2 := EvaluateDecision("claim-1", domain.EligibilitySources{
		DOPA: domain.DOPAResult{Valid: true, Alive: true, CardActive: true, Age: 17},
		KTB:  domain.KTBResult{PromptPayLinked: true},
	}, nil)
	if r2.Status != domain.StatusRejected {
		t.Errorf("expected Rejected for underage, got %s", r2.Status)
	}

	// Rule: SSO Section 33
	r3 := EvaluateDecision("claim-1", domain.EligibilitySources{
		DOPA: domain.DOPAResult{Valid: true, Alive: true, CardActive: true, Age: 25},
		SSO:  domain.SSOResult{Section: "33"},
		KTB:  domain.KTBResult{PromptPayLinked: true},
	}, nil)
	if r3.Status != domain.StatusRejected {
		t.Errorf("expected Rejected for SSO Section 33, got %s", r3.Status)
	}


	// Rule: PromptPay not linked
	r4 := EvaluateDecision("claim-1", domain.EligibilitySources{
		DOPA: domain.DOPAResult{Valid: true, Alive: true, CardActive: true, Age: 25},
		SSO:  domain.SSOResult{Section: "40"},
		KTB:  domain.KTBResult{PromptPayLinked: false},
	}, nil)
	if r4.Status != domain.StatusPending {
		t.Errorf("expected Pending for PromptPay not linked, got %s", r4.Status)
	}

	// Rule: Income too high
	r5 := EvaluateDecision("claim-1", domain.EligibilitySources{
		DOPA: domain.DOPAResult{Valid: true, Alive: true, CardActive: true, Age: 25},
		SSO:  domain.SSOResult{Section: "40"},
		KTB:  domain.KTBResult{PromptPayLinked: true, AverageMonthlyIncome: 35000},
	}, nil)
	if r5.Status != domain.StatusRejected {
		t.Errorf("expected Rejected for high income, got %s", r5.Status)
	}

	// decisionMessage checks
	m1 := decisionMessage(domain.DecisionResult{Status: domain.StatusApproved})
	if m1 != "claim approved automatically" {
		t.Errorf("expected claim approved automatically, got %s", m1)
	}

	m2 := decisionMessage(domain.DecisionResult{Status: domain.StatusRejected})
	if m2 != "claim rejected by eligibility rules" {
		t.Errorf("expected claim rejected by eligibility rules, got %s", m2)
	}

	m3 := decisionMessage(domain.DecisionResult{Status: domain.StatusPending})
	if m3 != "claim requires follow-up" {
		t.Errorf("expected claim requires follow-up, got %s", m3)
	}
}


