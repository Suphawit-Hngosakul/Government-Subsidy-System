package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"government-subsidy-system/backend/adapter"
	"government-subsidy-system/backend/controller"
	"government-subsidy-system/backend/repository"
	"government-subsidy-system/backend/service"

	_ "github.com/lib/pq"
)

func main() {
	// --- Auth & eKYC ---
	citizenRepo := repository.NewMemoryCitizenRepository()
	tokenRepo := repository.NewMemoryTokenRepository()
	authService := service.NewAuthService(citizenRepo, tokenRepo, envOrDefault("GEMINI_API_KEY", ""))
	providerDB := openProviderDatabase()
	defer providerDB.Close()
	providerService := service.NewProviderService(repository.NewPostgresProviderRepository(providerDB))

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
		providerService,
		providerService,
		providerService,
	)

	// --- Single mux, all routes ---
	mux := http.NewServeMux()
	controller.NewAuthHandler(authService).RegisterRoutes(mux)
	controller.NewAdminProjectHandler(projectService).RegisterRoutes(mux)
	controller.NewOfficerClaimHandler(officerService).RegisterRoutes(mux)
	controller.NewAdminDashboardHandler(dashboardService).RegisterRoutes(mux)
	controller.NewProviderHandler(providerService).RegisterRoutes(mux)
	controller.NewOrchestratorHTTPHandler(orchestratorService).RegisterRoutes(mux)

	log.Printf("government subsidy backend listening on :8080 (orchestrator=%s)", orchestratorURL)
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

func openProviderDatabase() *sql.DB {
	dsn := envOrDefault("PROVIDER_DATABASE_URL", "postgres://gss_user:gss_password@localhost:5433/gss_provider?sslmode=disable")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open provider database: %v", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("connect provider database: %v", err)
	}

	log.Printf("provider database connected")
	return db
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
