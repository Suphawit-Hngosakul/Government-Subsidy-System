package service

import (
	"context"
	"testing"

	"government-subsidy-system/backend/domain"
	"government-subsidy-system/backend/repository"
)

func newDashboardServiceForTest() (*DashboardService, *repository.MemoryProjectRepository, *repository.MemoryOfficerClaimRepository, *repository.MemoryAuditRepository) {
	projectRepo := repository.NewMemoryProjectRepository()
	claimRepo := repository.NewMemoryOfficerClaimRepository()
	auditRepo := repository.NewMemoryAuditRepository()
	return NewDashboardService(projectRepo, claimRepo, auditRepo), projectRepo, claimRepo, auditRepo
}

func TestDashboardServiceStatsCountsClaimsByStatus(t *testing.T) {
	dash, _, claimRepo, _ := newDashboardServiceForTest()
	ctx := context.Background()

	approveSeeded, _ := claimRepo.Get(ctx, "claim-001")
	approveSeeded.Status = domain.OfficerStatusApproved
	if _, err := claimRepo.Update(ctx, approveSeeded); err != nil {
		t.Fatalf("seed update failed: %v", err)
	}

	rejectSeeded, _ := claimRepo.Get(ctx, "claim-002")
	rejectSeeded.Status = domain.OfficerStatusRejected
	if _, err := claimRepo.Update(ctx, rejectSeeded); err != nil {
		t.Fatalf("seed update failed: %v", err)
	}

	stats := dash.Stats(ctx)
	if stats.Claims.Total != 3 {
		t.Fatalf("expected 3 total claims, got %d", stats.Claims.Total)
	}
	if stats.Claims.Approved != 1 || stats.Claims.Rejected != 1 || stats.Claims.Pending != 1 {
		t.Fatalf("unexpected claim breakdown: %+v", stats.Claims)
	}
}

func TestDashboardServiceStatsCountsActiveProjects(t *testing.T) {
	dash, projectRepo, _, _ := newDashboardServiceForTest()
	ctx := context.Background()

	if _, err := projectRepo.Create(ctx, domain.Project{ID: "p-1", Name: "Active", Active: true}); err != nil {
		t.Fatalf("seed active project failed: %v", err)
	}
	if _, err := projectRepo.Create(ctx, domain.Project{ID: "p-2", Name: "Inactive", Active: false}); err != nil {
		t.Fatalf("seed inactive project failed: %v", err)
	}

	stats := dash.Stats(ctx)
	if stats.Projects.Total != 2 || stats.Projects.Active != 1 {
		t.Fatalf("expected 2 total / 1 active, got %+v", stats.Projects)
	}
}

func TestDashboardServiceAuditLogReturnsNewestFirst(t *testing.T) {
	dash, _, _, auditRepo := newDashboardServiceForTest()
	ctx := context.Background()

	if _, err := auditRepo.Append(ctx, domain.AuditEntry{Actor: "officer-1", Action: domain.AuditActionClaimApproved, EntityID: "claim-001"}); err != nil {
		t.Fatalf("append 1 failed: %v", err)
	}
	if _, err := auditRepo.Append(ctx, domain.AuditEntry{Actor: "officer-2", Action: domain.AuditActionClaimRejected, EntityID: "claim-002"}); err != nil {
		t.Fatalf("append 2 failed: %v", err)
	}

	entries := dash.AuditLog(ctx, 10)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].EntityID != "claim-002" {
		t.Fatalf("expected newest first (claim-002), got %s", entries[0].EntityID)
	}
}

func TestDashboardServiceAuditLogRespectsLimit(t *testing.T) {
	dash, _, _, auditRepo := newDashboardServiceForTest()
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		if _, err := auditRepo.Append(ctx, domain.AuditEntry{Actor: "officer-1", Action: domain.AuditActionClaimApproved, EntityID: "claim-x"}); err != nil {
			t.Fatalf("append failed: %v", err)
		}
	}

	if entries := dash.AuditLog(ctx, 2); len(entries) != 2 {
		t.Fatalf("expected limit 2, got %d", len(entries))
	}
}

func TestDashboardServiceAuditLogAppliesDefaultAndMaxLimits(t *testing.T) {
	dash, _, _, auditRepo := newDashboardServiceForTest()
	ctx := context.Background()

	for i := 0; i < MaxAuditLimit+10; i++ {
		if _, err := auditRepo.Append(ctx, domain.AuditEntry{Actor: "x", Action: domain.AuditActionClaimApproved, EntityID: "y"}); err != nil {
			t.Fatalf("append failed: %v", err)
		}
	}

	if entries := dash.AuditLog(ctx, 0); len(entries) != DefaultAuditLimit {
		t.Fatalf("expected default limit %d, got %d", DefaultAuditLimit, len(entries))
	}
	if entries := dash.AuditLog(ctx, 9999); len(entries) != MaxAuditLimit {
		t.Fatalf("expected capped at %d, got %d", MaxAuditLimit, len(entries))
	}
}
