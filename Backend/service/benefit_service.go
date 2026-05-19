package service

import (
	"context"
	"errors"

	"government-subsidy-system/backend/domain"
	"government-subsidy-system/backend/repository"
)

type OrchestratorAdapter interface {
	Orchestrate(ctx context.Context, claimID string, nationalID string, projectID string) error
	GetDecision(ctx context.Context, claimID string) (domain.EligibilityResult, error)
}

type BenefitService struct {
	claimRepo           repository.BenefitClaimRepository
	projectRepo         repository.ProjectRepository
	orchestratorAdapter OrchestratorAdapter
}

func NewBenefitService(
	claimRepo repository.BenefitClaimRepository,
	projectRepo repository.ProjectRepository,
	orchestratorAdapter OrchestratorAdapter,
) *BenefitService {
	return &BenefitService{
		claimRepo:           claimRepo,
		projectRepo:         projectRepo,
		orchestratorAdapter: orchestratorAdapter,
	}
}

// SubmitClaim สร้างคำร้องขอสิทธิ์และ trigger orchestrator
func (s *BenefitService) SubmitClaim(ctx context.Context, req domain.BenefitClaimRequest) (domain.ClaimResponse, error) {
	if req.NationalID == "" {
		return domain.ClaimResponse{}, errors.New("nationalId is required")
	}
	if req.ProjectID == "" {
		return domain.ClaimResponse{}, errors.New("projectId is required")
	}

	// Verify project exists
	project, ok := s.projectRepo.Get(ctx, req.ProjectID)
	if !ok {
		return domain.ClaimResponse{}, errors.New("project not found")
	}

	// Create claim in repository
	newClaim := domain.Claim{
		NationalID: req.NationalID,
		ProjectID:  req.ProjectID,
		Status:     domain.StatusProcessing,
	}

	claim, err := s.claimRepo.Create(ctx, newClaim)
	if err != nil {
		return domain.ClaimResponse{}, err
	}

	// Trigger orchestrator async (fire and forget)
	go func() {
		_ = s.orchestratorAdapter.Orchestrate(context.Background(), claim.ID, req.NationalID, req.ProjectID)
	}()

	// Return created claim
	return domain.ClaimResponse{
		ID:          claim.ID,
		NationalID:  claim.NationalID,
		ProjectID:   claim.ProjectID,
		Status:      claim.Status,
		SubmittedAt: claim.SubmittedAt,
		UpdatedAt:   claim.UpdatedAt,
		Project:     &project,
	}, nil
}

// GetClaimStatus เช็คสถานะการพิจารณาของคำร้อง
func (s *BenefitService) GetClaimStatus(ctx context.Context, claimID string) (domain.ClaimResponse, error) {
	if claimID == "" {
		return domain.ClaimResponse{}, errors.New("claimId is required")
	}

	claim, err := s.claimRepo.GetByID(ctx, claimID)
	if err != nil {
		return domain.ClaimResponse{}, err
	}

	// Fetch latest decision from orchestrator
	eligibility, _ := s.orchestratorAdapter.GetDecision(ctx, claimID)
	if eligibility.ClaimID != "" {
		// Update eligibility if found
		_ = s.claimRepo.UpdateEligibility(ctx, claimID, eligibility)
		claim.Eligibility = &eligibility
		claim.Status = eligibility.Status
	}

	// Fetch project details
	project, ok := s.projectRepo.Get(ctx, claim.ProjectID)
	if ok {
		claim.Project = &project
	}

	return domain.ClaimResponse{
		ID:          claim.ID,
		NationalID:  claim.NationalID,
		ProjectID:   claim.ProjectID,
		Status:      claim.Status,
		SubmittedAt: claim.SubmittedAt,
		UpdatedAt:   claim.UpdatedAt,
		Eligibility: claim.Eligibility,
		Project:     claim.Project,
	}, nil
}

// GetClaimHistory ดูประวัติคำร้องทั้งหมดของประชาชนคนนี้
func (s *BenefitService) GetClaimHistory(ctx context.Context, nationalID string) ([]domain.ClaimResponse, error) {
	if nationalID == "" {
		return nil, errors.New("nationalId is required")
	}

	claims, err := s.claimRepo.GetByCitizen(ctx, nationalID)
	if err != nil {
		return nil, err
	}

	var responses []domain.ClaimResponse
	for _, claim := range claims {
		// Fetch latest decision from orchestrator
		eligibility, _ := s.orchestratorAdapter.GetDecision(ctx, claim.ID)
		if eligibility.ClaimID != "" {
			claim.Eligibility = &eligibility
			claim.Status = eligibility.Status
		}

		// Fetch project details
		project, ok := s.projectRepo.Get(ctx, claim.ProjectID)
		if !ok {
			project = domain.Project{} // Empty if project not found
		}

		responses = append(responses, domain.ClaimResponse{
			ID:          claim.ID,
			NationalID:  claim.NationalID,
			ProjectID:   claim.ProjectID,
			Status:      claim.Status,
			SubmittedAt: claim.SubmittedAt,
			UpdatedAt:   claim.UpdatedAt,
			Eligibility: claim.Eligibility,
			Project:     &project,
		})
	}

	return responses, nil
}

// GetAvailableProjects ดูรายการโครงการที่เปิดรับสิทธิ์
func (s *BenefitService) GetAvailableProjects(ctx context.Context) []domain.Project {
	allProjects := s.projectRepo.List(ctx)

	// Filter only active projects
	var activeProjects []domain.Project
	for _, p := range allProjects {
		if p.Active {
			activeProjects = append(activeProjects, p)
		}
	}

	return activeProjects
}
