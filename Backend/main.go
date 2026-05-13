package main

import (
	"log"
	"net/http"

	"government-subsidy-system/backend/controller"
	"government-subsidy-system/backend/repository"
	"government-subsidy-system/backend/service"
)

func main() {
	projectRepo := repository.NewMemoryProjectRepository()
	projectService := service.NewProjectService(projectRepo)

	mux := http.NewServeMux()
	controller.NewAdminProjectHandler(projectService).RegisterRoutes(mux)

	log.Println("government subsidy backend listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
