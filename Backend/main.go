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
	// --- Admin / Officer repos & services ---
	projectRepo := repository.NewMemoryProjectRepository()
	claimRepo := repository.NewMemoryOfficerClaimRepository()
	auditRepo := repository.NewMemoryAuditRepository()

	projectService := service.NewProjectService(projectRepo)

	orchestratorURL := envOrDefault("ORCHESTRATOR_BASE_URL", "http://localhost:8080")
	orchestratorClient := adapter.NewHTTPOrchestratorAdapter(orchestratorURL)
	officerService := service.NewOfficerService(claimRepo, orchestratorClient, auditRepo)
	dashboardService := service.NewDashboardService(projectRepo, claimRepo, auditRepo)

	// --- Orchestrator repos & service ---
	orchestratorRepo := repository.NewMemoryClaimRepository()
	orchestratorService := service.NewOrchestratorService(
		orchestratorRepo,
		adapter.NewMockDOPAAdapter(),
		adapter.NewMockSSOAdapter(),
		adapter.NewMockKTBAdapter(),
	)

	// --- Single mux, all routes ---
	mux := http.NewServeMux()
	controller.NewAdminProjectHandler(projectService).RegisterRoutes(mux)
	controller.NewOfficerClaimHandler(officerService).RegisterRoutes(mux)
	controller.NewAdminDashboardHandler(dashboardService).RegisterRoutes(mux)
	controller.NewOrchestratorHTTPHandler(orchestratorService).RegisterRoutes(mux)

	log.Printf("government subsidy backend listening on :8080 (orchestrator=%s)", orchestratorURL)
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
