package service

import (
	"context"
	"testing"

	"government-subsidy-system/backend/adapter"
	"government-subsidy-system/backend/domain"
	"government-subsidy-system/backend/repository"
)

func TestSubmitClaimCreatesClaimAndTriggersOrchestrator(t *testing.T) {
	claimRepo := repository.NewMemoryBenefitClaimRepository()
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
	claimRepo := repository.NewMemoryBenefitClaimRepository()
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
	claimRepo := repository.NewMemoryBenefitClaimRepository()
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
	claimRepo := repository.NewMemoryBenefitClaimRepository()
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
	claimRepo := repository.NewMemoryBenefitClaimRepository()
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
	claimRepo := repository.NewMemoryBenefitClaimRepository()
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
