package service

import (
	"context"

	"government-subsidy-system/backend/domain"
	"government-subsidy-system/backend/repository"
)

const (
	DefaultAuditLimit = 50
	MaxAuditLimit     = 200
)

type AuditReader interface {
	List(ctx context.Context, limit int) []domain.AuditEntry
}

type DashboardService struct {
	projectRepo repository.ProjectRepository
	claimRepo   repository.OfficerClaimRepository
	auditRepo   AuditReader
}

func NewDashboardService(projectRepo repository.ProjectRepository, claimRepo repository.OfficerClaimRepository, auditRepo AuditReader) *DashboardService {
	return &DashboardService{projectRepo: projectRepo, claimRepo: claimRepo, auditRepo: auditRepo}
}

func (s *DashboardService) Stats(ctx context.Context) domain.Stats {
	projects := s.projectRepo.List(ctx)
	active := 0
	for _, p := range projects {
		if p.Active {
			active++
		}
	}

	pending := len(s.claimRepo.ListByStatus(ctx, domain.OfficerStatusPending))
	approved := len(s.claimRepo.ListByStatus(ctx, domain.OfficerStatusApproved))
	rejected := len(s.claimRepo.ListByStatus(ctx, domain.OfficerStatusRejected))

	return domain.Stats{
		Projects: domain.ProjectStats{Total: len(projects), Active: active},
		Claims: domain.ClaimStats{
			Total:    pending + approved + rejected,
			Pending:  pending,
			Approved: approved,
			Rejected: rejected,
		},
	}
}

func (s *DashboardService) AuditLog(ctx context.Context, limit int) []domain.AuditEntry {
	if limit <= 0 {
		limit = DefaultAuditLimit
	}
	if limit > MaxAuditLimit {
		limit = MaxAuditLimit
	}
	return s.auditRepo.List(ctx, limit)
}
