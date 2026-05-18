package main

import (
	"log"
	"net/http"
	"os"

	"government-subsidy-system/backend/adapter"
	"government-subsidy-system/backend/controller"
	"government-subsidy-system/backend/repository"
	"government-subsidy-system/backend/service"
)

func main() {
	projectRepo := repository.NewMemoryProjectRepository()
	claimRepo := repository.NewMemoryOfficerClaimRepository()
	auditRepo := repository.NewMemoryAuditRepository()

	projectService := service.NewProjectService(projectRepo)

	orchestratorURL := envOrDefault("ORCHESTRATOR_BASE_URL", "http://localhost:8080")
	orchestratorClient := adapter.NewHTTPOrchestratorAdapter(orchestratorURL)
	officerService := service.NewOfficerService(claimRepo, orchestratorClient, auditRepo)

	dashboardService := service.NewDashboardService(projectRepo, claimRepo, auditRepo)

	mux := http.NewServeMux()
	controller.NewAdminProjectHandler(projectService).RegisterRoutes(mux)
	controller.NewOfficerClaimHandler(officerService).RegisterRoutes(mux)
	controller.NewAdminDashboardHandler(dashboardService).RegisterRoutes(mux)

	log.Printf("government subsidy backend listening on :8080 (orchestrator=%s)", orchestratorURL)
	repo := repository.NewMemoryClaimRepository()
	orchestrator := service.NewOrchestratorService(
		repo,
		adapter.NewMockDOPAAdapter(),
		adapter.NewMockSSOAdapter(),
		adapter.NewMockKTBAdapter(),
	)

	mux := http.NewServeMux()
	controller.NewOrchestratorHTTPHandler(orchestrator).RegisterRoutes(mux)

	log.Println("government subsidy backend listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
