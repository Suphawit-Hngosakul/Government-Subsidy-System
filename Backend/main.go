package main

import (
	"log"
	"net/http"

	"government-subsidy-system/backend/controller"
	"government-subsidy-system/backend/repository"
	"government-subsidy-system/backend/service"
)

func main() {
	repo := repository.NewMemoryClaimRepository()
	orchestrator := service.NewOrchestratorService(
		repo,
		service.MockDOPAClient{},
		service.MockSSOClient{},
		service.MockKTBClient{},
	)

	mux := http.NewServeMux()
	controller.NewOrchestratorHTTPHandler(orchestrator).RegisterRoutes(mux)

	log.Println("government subsidy backend listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
