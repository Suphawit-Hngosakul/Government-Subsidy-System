package service

import (
	"context"
	"errors"
	"testing"

	"government-subsidy-system/backend/adapter"
	"government-subsidy-system/backend/domain"
	"government-subsidy-system/backend/repository"
)

type errorSSOClient struct{}

func (errorSSOClient) Status(ctx context.Context, nationalID string) (domain.SSOResult, error) {
	return domain.SSOResult{}, errors.New("SSO down")
}

type errorKTBClient struct{}

func (errorKTBClient) FinancialCheck(ctx context.Context, nationalID string) (domain.KTBResult, error) {
	return domain.KTBResult{}, errors.New("KTB down")
}

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
	projectRepo := repository.NewMemoryProjectRepository()
	svc := NewOrchestratorService(
		repo,
		projectRepo,
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
	projectRepo := repository.NewMemoryProjectRepository()
	svc := NewOrchestratorService(
		repo,
		projectRepo,
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

// TestOrchestrateSelectiveQueries verifies that SSO and KTB integrations are skipped if not required,
// avoiding errors even if their downstream service adaptors fail.
func TestOrchestrateSelectiveQueries(t *testing.T) {
	repo := repository.NewMemoryClaimRepository()
	projectRepo := repository.NewMemoryProjectRepository()

	// 1. Setup a project that does NOT require SSO or KTB checks
	ctx := context.Background()
	_, _ = projectRepo.Create(ctx, domain.Project{
		ID:     "proj-selective",
		Name:   "Selective Subsidy",
		Active: true,
		Criteria: domain.ProjectCriteria{
			MinAge:           18,
			RequirePromptPay: false,
			RequireSSO:       false,
			RequireKTB:       false,
		},
	})

	// 2. Setup service with error-prone SSO and KTB clients
	svc := NewOrchestratorService(
		repo,
		projectRepo,
		adapter.NewMockDOPAAdapter(),
		errorSSOClient{},
		errorKTBClient{},
	)

	// 3. Orchestrate with selective project criteria
	result, err := svc.Orchestrate(ctx, domain.OrchestrateRequest{
		ClaimID:    "claim-selective",
		NationalID: "1101700203451",
		ProjectID:  "proj-selective",
	})
	if err != nil {
		t.Fatalf("orchestrate returned error: %v", err)
	}

	// 4. Assert orchestration succeeds and remains Approved (since error-prone clients were bypassed)
	if result.Status != domain.StatusApproved {
		t.Fatalf("expected status %s since SSO/KTB were bypassed, got %s (reasons: %v)", domain.StatusApproved, result.Status, result.Reasons)
	}
}

// TestEvaluateDecisionWithCustomCriteria verifies diverse and customizable rule limits:
// 1. Custom age boundaries (e.g. MinAge: 16, MaxAge: 60)
// 2. Custom financial limits (e.g. MaxDepositTotal: 400000, MaxMonthlyIncome: 50000)
// 3. Whitelisted SSO sections (e.g. AllowedSSOSections: ["39", "40"])
func TestEvaluateDecisionWithCustomCriteria(t *testing.T) {
	criteria := domain.ProjectCriteria{
		MinAge:           16,
		MaxAge:           60,
		RequirePromptPay: true,
		RequireSSO:       true,
		RequireKTB:       true,
		MaxMonthlyIncome: 50000,
		MaxDepositTotal:  400000,
		AllowedSSOSections: []string{"39", "40"},
	}

	// Case 1: Approved - satisfies all custom criteria
	sources1 := domain.EligibilitySources{
		DOPA: domain.DOPAResult{Valid: true, Age: 16, Alive: true, CardActive: true},
		SSO:  domain.SSOResult{Section: "40", ContributionMonths: 12},
		KTB:  domain.KTBResult{DepositTotal: 399999, AverageMonthlyIncome: 45000, PromptPayLinked: true},
	}
	r1 := EvaluateDecisionWithCriteria("claim-c1", sources1, nil, criteria)
	if r1.Status != domain.StatusApproved {
		t.Errorf("expected Approved, got %s (reasons: %v)", r1.Status, r1.Reasons)
	}

	// Case 2: Rejected - Under 16 (Age 15)
	sources2 := domain.EligibilitySources{
		DOPA: domain.DOPAResult{Valid: true, Age: 15, Alive: true, CardActive: true},
		SSO:  domain.SSOResult{Section: "40"},
		KTB:  domain.KTBResult{DepositTotal: 10000, AverageMonthlyIncome: 10000, PromptPayLinked: true},
	}
	r2 := EvaluateDecisionWithCriteria("claim-c2", sources2, nil, criteria)
	if r2.Status != domain.StatusRejected {
		t.Errorf("expected Rejected, got %s", r2.Status)
	}

	// Case 3: Rejected - Over 60 (Age 61)
	sources3 := domain.EligibilitySources{
		DOPA: domain.DOPAResult{Valid: true, Age: 61, Alive: true, CardActive: true},
		SSO:  domain.SSOResult{Section: "40"},
		KTB:  domain.KTBResult{DepositTotal: 10000, AverageMonthlyIncome: 10000, PromptPayLinked: true},
	}
	r3 := EvaluateDecisionWithCriteria("claim-c3", sources3, nil, criteria)
	if r3.Status != domain.StatusRejected {
		t.Errorf("expected Rejected, got %s", r3.Status)
	}

	// Case 4: Rejected - Deposit total exceeds 400,000 THB (Deposit: 400,001 THB)
	sources4 := domain.EligibilitySources{
		DOPA: domain.DOPAResult{Valid: true, Age: 25, Alive: true, CardActive: true},
		SSO:  domain.SSOResult{Section: "40"},
		KTB:  domain.KTBResult{DepositTotal: 400001, AverageMonthlyIncome: 10000, PromptPayLinked: true},
	}
	r4 := EvaluateDecisionWithCriteria("claim-c4", sources4, nil, criteria)
	if r4.Status != domain.StatusRejected {
		t.Errorf("expected Rejected, got %s", r4.Status)
	}

	// Case 5: Rejected - SSO Section is not in whitelist (Section 33)
	sources5 := domain.EligibilitySources{
		DOPA: domain.DOPAResult{Valid: true, Age: 25, Alive: true, CardActive: true},
		SSO:  domain.SSOResult{Section: "33"},
		KTB:  domain.KTBResult{DepositTotal: 10000, AverageMonthlyIncome: 10000, PromptPayLinked: true},
	}
	r5 := EvaluateDecisionWithCriteria("claim-c5", sources5, nil, criteria)
	if r5.Status != domain.StatusRejected {
		t.Errorf("expected Rejected, got %s", r5.Status)
	}
}


