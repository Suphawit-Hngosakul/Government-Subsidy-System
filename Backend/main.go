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
	projectService := service.NewProjectService(projectRepo)

	claimRepo := repository.NewMemoryOfficerClaimRepository()
	orchestratorURL := envOrDefault("ORCHESTRATOR_BASE_URL", "http://localhost:8080")
	orchestratorClient := adapter.NewHTTPOrchestratorAdapter(orchestratorURL)
	officerService := service.NewOfficerService(claimRepo, orchestratorClient)

	mux := http.NewServeMux()
	controller.NewAdminProjectHandler(projectService).RegisterRoutes(mux)
	controller.NewOfficerClaimHandler(officerService).RegisterRoutes(mux)

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
