package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"government-subsidy-system/backend/domain"
	"government-subsidy-system/backend/repository"
)

type stubOrchestrator struct {
	decisions map[string]domain.EligibilityResult
	err       error
}

func (s *stubOrchestrator) GetDecision(_ context.Context, claimID string) (domain.EligibilityResult, error) {
	if s.err != nil {
		return domain.EligibilityResult{}, s.err
	}
	if d, ok := s.decisions[claimID]; ok {
		return d, nil
	}
	return domain.EligibilityResult{}, nil
}

func newOfficerServiceForTest(orch OrchestratorClient) (*OfficerService, *repository.MemoryOfficerClaimRepository, *repository.MemoryAuditRepository) {
	repo := repository.NewMemoryOfficerClaimRepository()
	audit := repository.NewMemoryAuditRepository()
	return NewOfficerService(repo, orch, audit), repo, audit
}

func TestOfficerServiceListPendingReturnsSeededClaims(t *testing.T) {
	svc, _, _ := newOfficerServiceForTest(&stubOrchestrator{})

	list := svc.ListPending(context.Background())
	if len(list) < 1 {
		t.Fatalf("expected at least one seeded pending claim, got %d", len(list))
	}
	for _, c := range list {
		if c.Status != domain.OfficerStatusPending {
			t.Fatalf("expected pending claim, got %s", c.Status)
		}
	}
}

func TestOfficerServiceDetailMergesClaimAndEligibility(t *testing.T) {
	expected := domain.EligibilityResult{
		ClaimID: "claim-001",
		Status:  "approved",
		Reasons: []string{"eligible under project criteria rules"},
		Sources: domain.EligibilitySources{
			DOPA: domain.EligibilityDOPA{Valid: true, Age: 35, Alive: true, CardActive: true},
			SSO:  domain.EligibilitySSO{Section: "40", ContributionMonths: 12},
			KTB:  domain.EligibilityKTB{DepositTotal: 12000, AverageMonthlyIncome: 15000, PromptPayLinked: true},
		},
	}
	orch := &stubOrchestrator{decisions: map[string]domain.EligibilityResult{"claim-001": expected}}
	svc, _, _ := newOfficerServiceForTest(orch)

	detail, err := svc.Detail(context.Background(), "claim-001")
	if err != nil {
		t.Fatalf("detail returned error: %v", err)
	}
	if detail.Claim.ClaimID != "claim-001" {
		t.Fatalf("expected claim-001, got %s", detail.Claim.ClaimID)
	}
	if detail.Eligibility == nil {
		t.Fatal("expected eligibility to be populated")
	}
	if detail.Eligibility.Status != "approved" {
		t.Fatalf("expected eligibility status approved, got %s", detail.Eligibility.Status)
	}
	if detail.Eligibility.Sources.DOPA.Age != 35 {
		t.Fatalf("expected DOPA age 35, got %d", detail.Eligibility.Sources.DOPA.Age)
	}
}

func TestOfficerServiceDetailUnknownClaimReturnsNotFound(t *testing.T) {
	svc, _, _ := newOfficerServiceForTest(&stubOrchestrator{})

	_, err := svc.Detail(context.Background(), "claim-missing")
	if !errors.Is(err, repository.ErrOfficerClaimNotFound) {
		t.Fatalf("expected ErrClaimNotFound, got %v", err)
	}
}

func TestOfficerServiceDetailReturnsNilEligibilityWhenOrchestratorHasNoDecision(t *testing.T) {
	svc, _, _ := newOfficerServiceForTest(&stubOrchestrator{})

	detail, err := svc.Detail(context.Background(), "claim-001")
	if err != nil {
		t.Fatalf("detail returned error: %v", err)
	}
	if detail.Eligibility != nil {
		t.Fatalf("expected eligibility to be nil, got %+v", detail.Eligibility)
	}
}

func TestOfficerServiceDetailSwallowsOrchestratorErrorAndReturnsClaim(t *testing.T) {
	boom := errors.New("network down")
	svc, _, _ := newOfficerServiceForTest(&stubOrchestrator{err: boom})

	detail, err := svc.Detail(context.Background(), "claim-001")
	if err != nil {
		t.Fatalf("expected no error when orchestrator fails, got %v", err)
	}
	if detail.Claim.ClaimID != "claim-001" {
		t.Fatalf("expected claim-001 still returned, got %q", detail.Claim.ClaimID)
	}
	if detail.Eligibility != nil {
		t.Fatal("expected eligibility to be nil when orchestrator errors")
	}
}

func TestOfficerServiceDetailEmptyIDReturnsErrClaimIDRequired(t *testing.T) {
	svc, _, _ := newOfficerServiceForTest(&stubOrchestrator{})

	_, err := svc.Detail(context.Background(), "")
	if !errors.Is(err, ErrClaimIDRequired) {
		t.Fatalf("expected ErrClaimIDRequired, got %v", err)
	}
}

func TestOfficerServiceDecideEmptyIDReturnsErrClaimIDRequired(t *testing.T) {
	svc, _, _ := newOfficerServiceForTest(&stubOrchestrator{})

	if _, err := svc.Approve(context.Background(), "", domain.OfficerDecisionInput{OfficerID: "officer-1"}); !errors.Is(err, ErrClaimIDRequired) {
		t.Fatalf("Approve(\"\"): expected ErrClaimIDRequired, got %v", err)
	}
	if _, err := svc.Reject(context.Background(), "", domain.OfficerDecisionInput{OfficerID: "officer-1", Reason: "x"}); !errors.Is(err, ErrClaimIDRequired) {
		t.Fatalf("Reject(\"\"): expected ErrClaimIDRequired, got %v", err)
	}
}

func TestOfficerServiceApproveTransitionsPendingToApproved(t *testing.T) {
	svc, _, _ := newOfficerServiceForTest(&stubOrchestrator{})

	claim, err := svc.Approve(context.Background(), "claim-001", domain.OfficerDecisionInput{OfficerID: "officer-1"})
	if err != nil {
		t.Fatalf("approve returned error: %v", err)
	}
	if claim.Status != domain.OfficerStatusApproved {
		t.Fatalf("expected approved, got %s", claim.Status)
	}
	if claim.OfficerDecision == nil || claim.OfficerDecision.OfficerID != "officer-1" {
		t.Fatal("expected officer decision recorded")
	}
	if time.Since(claim.OfficerDecision.DecidedAt) > time.Minute {
		t.Fatal("expected DecidedAt to be recent")
	}
}

func TestOfficerServiceApproveRequiresOfficerID(t *testing.T) {
	svc, _, _ := newOfficerServiceForTest(&stubOrchestrator{})

	_, err := svc.Approve(context.Background(), "claim-001", domain.OfficerDecisionInput{})
	if !errors.Is(err, ErrOfficerIDRequired) {
		t.Fatalf("expected ErrOfficerIDRequired, got %v", err)
	}
}

func TestOfficerServiceRejectRequiresReason(t *testing.T) {
	svc, _, _ := newOfficerServiceForTest(&stubOrchestrator{})

	_, err := svc.Reject(context.Background(), "claim-001", domain.OfficerDecisionInput{OfficerID: "officer-1"})
	if !errors.Is(err, ErrRejectReasonRequired) {
		t.Fatalf("expected ErrRejectReasonRequired, got %v", err)
	}
}

func TestOfficerServiceRejectTransitionsPendingToRejected(t *testing.T) {
	svc, _, _ := newOfficerServiceForTest(&stubOrchestrator{})

	claim, err := svc.Reject(context.Background(), "claim-001", domain.OfficerDecisionInput{
		OfficerID: "officer-2",
		Reason:    "income exceeds threshold",
	})
	if err != nil {
		t.Fatalf("reject returned error: %v", err)
	}
	if claim.Status != domain.OfficerStatusRejected {
		t.Fatalf("expected rejected, got %s", claim.Status)
	}
	if claim.OfficerDecision == nil || claim.OfficerDecision.Reason != "income exceeds threshold" {
		t.Fatal("expected officer reason recorded")
	}
}

func TestOfficerServiceCannotDecideAlreadyDecidedClaim(t *testing.T) {
	svc, _, _ := newOfficerServiceForTest(&stubOrchestrator{})
	ctx := context.Background()

	if _, err := svc.Approve(ctx, "claim-001", domain.OfficerDecisionInput{OfficerID: "officer-1"}); err != nil {
		t.Fatalf("seed approve failed: %v", err)
	}

	_, err := svc.Reject(ctx, "claim-001", domain.OfficerDecisionInput{OfficerID: "officer-2", Reason: "changed mind"})
	if !errors.Is(err, ErrClaimAlreadyDecided) {
		t.Fatalf("expected ErrClaimAlreadyDecided, got %v", err)
	}
}

func TestOfficerServiceUnknownClaimDecisionReturnsNotFound(t *testing.T) {
	svc, _, _ := newOfficerServiceForTest(&stubOrchestrator{})

	_, err := svc.Approve(context.Background(), "claim-missing", domain.OfficerDecisionInput{OfficerID: "officer-1"})
	if !errors.Is(err, repository.ErrOfficerClaimNotFound) {
		t.Fatalf("expected ErrClaimNotFound, got %v", err)
	}
}

func TestOfficerServiceApproveAppendsAuditEntry(t *testing.T) {
	svc, _, audit := newOfficerServiceForTest(&stubOrchestrator{})
	ctx := context.Background()

	if _, err := svc.Approve(ctx, "claim-001", domain.OfficerDecisionInput{OfficerID: "officer-1", Reason: "verified"}); err != nil {
		t.Fatalf("approve returned error: %v", err)
	}

	entries := audit.List(ctx, 10)
	if len(entries) != 1 {
		t.Fatalf("expected 1 audit entry, got %d", len(entries))
	}
	got := entries[0]
	if got.Action != domain.AuditActionClaimApproved {
		t.Fatalf("expected action %s, got %s", domain.AuditActionClaimApproved, got.Action)
	}
	if got.Actor != "officer-1" || got.EntityID != "claim-001" {
		t.Fatalf("unexpected actor/entity in audit: %+v", got)
	}
	if got.Metadata["reason"] != "verified" {
		t.Fatalf("expected reason in metadata, got %+v", got.Metadata)
	}
}

func TestOfficerServiceRejectAppendsAuditEntry(t *testing.T) {
	svc, _, audit := newOfficerServiceForTest(&stubOrchestrator{})
	ctx := context.Background()

	if _, err := svc.Reject(ctx, "claim-002", domain.OfficerDecisionInput{OfficerID: "officer-2", Reason: "income high"}); err != nil {
		t.Fatalf("reject returned error: %v", err)
	}

	entries := audit.List(ctx, 10)
	if len(entries) != 1 || entries[0].Action != domain.AuditActionClaimRejected {
		t.Fatalf("expected single claim.rejected audit entry, got %+v", entries)
	}
}

func TestOfficerServiceFailedDecisionDoesNotAppendAudit(t *testing.T) {
	svc, _, audit := newOfficerServiceForTest(&stubOrchestrator{})
	ctx := context.Background()

	_, _ = svc.Approve(ctx, "claim-missing", domain.OfficerDecisionInput{OfficerID: "officer-1"})

	if entries := audit.List(ctx, 10); len(entries) != 0 {
		t.Fatalf("expected no audit entries on failed decision, got %d", len(entries))
	}
}
