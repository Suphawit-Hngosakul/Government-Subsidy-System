package service

import (
	"context"
	"testing"

	"government-subsidy-system/backend/adapter"
	"government-subsidy-system/backend/domain"
	"government-subsidy-system/backend/repository"
)

func TestSubmitClaimCreatesClaimAndTriggersOrchestrator(t *testing.T) {
	claimRepo := repository.NewMemoryBenefitClaimRepository(nil)
	projectRepo := repository.NewMemoryProjectRepository()
	orchestratorClient := adapter.NewHTTPOrchestratorAdapter("http://localhost:8080")

	svc := NewBenefitService(claimRepo, projectRepo, orchestratorClient)
	projectSvc := NewProjectService(projectRepo)

	// Create a project first using ProjectService
	project, err := projectSvc.Create(context.Background(), domain.ProjectInput{
		Name:        "Test Project",
		Description: "Test Description",
		Active:      true,
		Criteria: domain.ProjectCriteria{
			MinAge:           18,
			RequirePromptPay: true,
		},
	})
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Submit claim
	resp, err := svc.SubmitClaim(context.Background(), domain.BenefitClaimRequest{
		NationalID: "1101700203451",
		ProjectID:  project.ID,
	})

	if err != nil {
		t.Fatalf("SubmitClaim returned error: %v", err)
	}
	if resp.NationalID != "1101700203451" {
		t.Fatalf("expected nationalId 1101700203451, got %s", resp.NationalID)
	}
	if resp.Status != domain.StatusProcessing {
		t.Fatalf("expected status %v, got %v", domain.StatusProcessing, resp.Status)
	}
	if resp.ID == "" {
		t.Fatal("expected claim ID to be generated")
	}
}

func TestSubmitClaimRejectsEmptyNationalID(t *testing.T) {
	claimRepo := repository.NewMemoryBenefitClaimRepository(nil)
	projectRepo := repository.NewMemoryProjectRepository()
	orchestratorClient := adapter.NewHTTPOrchestratorAdapter("http://localhost:8080")

	svc := NewBenefitService(claimRepo, projectRepo, orchestratorClient)

	_, err := svc.SubmitClaim(context.Background(), domain.BenefitClaimRequest{
		NationalID: "",
		ProjectID:  "project-1",
	})

	if err == nil {
		t.Fatal("expected error for empty nationalId")
	}
}

func TestSubmitClaimRejectsNonExistentProject(t *testing.T) {
	claimRepo := repository.NewMemoryBenefitClaimRepository(nil)
	projectRepo := repository.NewMemoryProjectRepository()
	orchestratorClient := adapter.NewHTTPOrchestratorAdapter("http://localhost:8080")

	svc := NewBenefitService(claimRepo, projectRepo, orchestratorClient)

	_, err := svc.SubmitClaim(context.Background(), domain.BenefitClaimRequest{
		NationalID: "1101700203451",
		ProjectID:  "nonexistent-project",
	})

	if err == nil {
		t.Fatal("expected error for non-existent project")
	}
	if err.Error() != "project not found" {
		t.Fatalf("expected 'project not found', got '%s'", err.Error())
	}
}

func TestGetClaimStatusReturnsClaim(t *testing.T) {
	claimRepo := repository.NewMemoryBenefitClaimRepository(nil)
	projectRepo := repository.NewMemoryProjectRepository()
	orchestratorClient := adapter.NewHTTPOrchestratorAdapter("http://localhost:8080")

	svc := NewBenefitService(claimRepo, projectRepo, orchestratorClient)
	projectSvc := NewProjectService(projectRepo)

	// Create a project and claim
	project, err := projectSvc.Create(context.Background(), domain.ProjectInput{
		Name:        "Test Project",
		Description: "Test Description",
		Active:      true,
		Criteria: domain.ProjectCriteria{
			MinAge:           18,
			RequirePromptPay: true,
		},
	})
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	resp, _ := svc.SubmitClaim(context.Background(), domain.BenefitClaimRequest{
		NationalID: "1101700203451",
		ProjectID:  project.ID,
	})

	// Get claim status
	status, err := svc.GetClaimStatus(context.Background(), resp.ID)
	if err != nil {
		t.Fatalf("GetClaimStatus returned error: %v", err)
	}
	if status.ID != resp.ID {
		t.Fatalf("expected claim ID %s, got %s", resp.ID, status.ID)
	}
	if status.Status != domain.StatusProcessing {
		t.Fatalf("expected status %v, got %v", domain.StatusProcessing, status.Status)
	}
}

func TestGetClaimHistoryReturnsCitizensAllClaims(t *testing.T) {
	claimRepo := repository.NewMemoryBenefitClaimRepository(nil)
	projectRepo := repository.NewMemoryProjectRepository()
	orchestratorClient := adapter.NewHTTPOrchestratorAdapter("http://localhost:8080")

	svc := NewBenefitService(claimRepo, projectRepo, orchestratorClient)
	projectSvc := NewProjectService(projectRepo)

	// Create a project
	project, err := projectSvc.Create(context.Background(), domain.ProjectInput{
		Name:        "Test Project",
		Description: "Test Description",
		Active:      true,
		Criteria: domain.ProjectCriteria{
			MinAge:           18,
			RequirePromptPay: true,
		},
	})
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Submit multiple claims for same citizen
	svc.SubmitClaim(context.Background(), domain.BenefitClaimRequest{
		NationalID: "1101700203451",
		ProjectID:  project.ID,
	})
	svc.SubmitClaim(context.Background(), domain.BenefitClaimRequest{
		NationalID: "1101700203451",
		ProjectID:  project.ID,
	})

	// Get history
	history, err := svc.GetClaimHistory(context.Background(), "1101700203451")
	if err != nil {
		t.Fatalf("GetClaimHistory returned error: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("expected 2 claims, got %d", len(history))
	}
}

func TestGetAvailableProjectsReturnOnlyActiveProjects(t *testing.T) {
	claimRepo := repository.NewMemoryBenefitClaimRepository(nil)
	projectRepo := repository.NewMemoryProjectRepository()
	orchestratorClient := adapter.NewHTTPOrchestratorAdapter("http://localhost:8080")

	svc := NewBenefitService(claimRepo, projectRepo, orchestratorClient)

	// Create active and inactive projects
	projectRepo.Create(context.Background(), domain.Project{
		Name:   "Active Project",
		Active: true,
	})
	projectRepo.Create(context.Background(), domain.Project{
		Name:   "Inactive Project",
		Active: false,
	})

	// Get available projects
	projects := svc.GetAvailableProjects(context.Background())
	if len(projects) != 1 {
		t.Fatalf("expected 1 active project, got %d", len(projects))
	}
	if projects[0].Name != "Active Project" {
		t.Fatalf("expected 'Active Project', got '%s'", projects[0].Name)
	}
}

func TestBenefitServiceErrors(t *testing.T) {
	claimRepo := repository.NewMemoryBenefitClaimRepository(nil)
	projectRepo := repository.NewMemoryProjectRepository()
	orchestratorClient := adapter.NewHTTPOrchestratorAdapter("http://localhost:8080")

	svc := NewBenefitService(claimRepo, projectRepo, orchestratorClient)
	ctx := context.Background()

	// SubmitClaim missing project ID
	_, err := svc.SubmitClaim(ctx, domain.BenefitClaimRequest{
		NationalID: "1101700203451",
		ProjectID:  "",
	})
	if err == nil || err.Error() != "projectId is required" {
		t.Errorf("expected 'projectId is required', got %v", err)
	}

	// GetClaimStatus missing claim ID
	_, err = svc.GetClaimStatus(ctx, "")
	if err == nil || err.Error() != "claimId is required" {
		t.Errorf("expected 'claimId is required', got %v", err)
	}

	// GetClaimStatus unknown claim ID (triggers repository error)
	_, err = svc.GetClaimStatus(ctx, "unknown-claim")
	if err == nil {
		t.Error("expected error for unknown claim ID, got nil")
	}

	// GetClaimHistory missing national ID
	_, err = svc.GetClaimHistory(ctx, "")
	if err == nil || err.Error() != "nationalId is required" {
		t.Errorf("expected 'nationalId is required', got %v", err)
	}
}

type mockOrchestratorAdapter struct {
	decision domain.EligibilityResult
}

func (m *mockOrchestratorAdapter) Orchestrate(ctx context.Context, claimID string, nationalID string, projectID string) error {
	return nil
}

func (m *mockOrchestratorAdapter) GetDecision(ctx context.Context, claimID string) (domain.EligibilityResult, error) {
	return m.decision, nil
}

func TestGetClaimStatusAndHistoryWithDecision(t *testing.T) {
	claimRepo := repository.NewMemoryBenefitClaimRepository(nil)
	projectRepo := repository.NewMemoryProjectRepository()
	
	decision := domain.EligibilityResult{
		ClaimID: "claim-123",
		Status:  domain.StatusApproved,
		Reasons: []string{"eligible"},
	}
	mockOrch := &mockOrchestratorAdapter{decision: decision}

	svc := NewBenefitService(claimRepo, projectRepo, mockOrch)
	projectSvc := NewProjectService(projectRepo)
	ctx := context.Background()

	// Create project and claim
	project, _ := projectSvc.Create(ctx, domain.ProjectInput{
		Name:   "Active Project",
		Active: true,
	})

	claim, err := claimRepo.Create(ctx, domain.Claim{
		ID:         "claim-123",
		NationalID: "1101700203451",
		ProjectID:  project.ID,
		Status:     domain.StatusProcessing,
	})
	if err != nil {
		t.Fatalf("failed to create claim: %v", err)
	}

	// 1. GetClaimStatus with valid decision
	resp, err := svc.GetClaimStatus(ctx, claim.ID)
	if err != nil {
		t.Fatalf("GetClaimStatus failed: %v", err)
	}
	if resp.Status != domain.StatusApproved {
		t.Errorf("expected status Approved, got %s", resp.Status)
	}

	// 2. GetClaimHistory with valid decision
	history, err := svc.GetClaimHistory(ctx, "1101700203451")
	if err != nil {
		t.Fatalf("GetClaimHistory failed: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("expected 1 claim in history, got %d", len(history))
	}
	if history[0].Status != domain.StatusApproved {
		t.Errorf("expected history claim status Approved, got %s", history[0].Status)
	}
}


