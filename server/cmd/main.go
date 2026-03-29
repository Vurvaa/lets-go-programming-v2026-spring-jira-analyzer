package main

import (
	"log"
	"net/http"
	"server/internals/handler"
	"time"

	"server/internals/repository/postgres"
	"server/internals/service"
)

func main() {
	time.Sleep(45 * time.Second)

	db := postgres.NewDB()
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to close db: %v", err)
		}
	}()

	projectRepo := postgres.NewProjectRepository(db)
	projectService := service.NewProjectService(projectRepo)
	projectHandler := handler.NewProjectHandler(projectService)

	projects, err := projectService.GetAllProjects(0, 10)
	if err != nil {
		log.Fatalf("failed to get projects: %v", err)
	}
	log.Printf("projects: %+v", projects)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/projects", projectHandler.GetAllProjects)
	mux.HandleFunc("GET /api/v1/projects/{id}", projectHandler.GetProjectByID)
	mux.HandleFunc("DELETE /api/v1/projects/{id}", projectHandler.DeleteProjectByID)

	log.Println("Starting API server on :8000")
	server := &http.Server{
		Addr:         ":8000",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("failed to start API server: %v", err)
	}
}
