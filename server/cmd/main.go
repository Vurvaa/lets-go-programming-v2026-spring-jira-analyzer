package main

import (
	"log"
	"net/http"
	"server/internals/repository/postgres"
	"time"

	"server/internals/handler"
	"server/internals/repository/connector"
	"server/internals/router"
	"server/internals/service"
)

func main() {
	const configName = "configs/config.yaml"

	db := postgres.NewDB(configName)
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to close db: %v", err)
		}
	}()

	projectRepo := postgres.NewDBRepository(db)
	connectorRepo := connector.NewConnectorRepository(configName)
	projectService := service.NewProjectService(connectorRepo, projectRepo)
	projectHandler := handler.NewProjectHandler(projectService)
	projectRouter := router.NewRouter(projectHandler)

	log.Println("Starting API server on :8000")
	server := &http.Server{
		Addr:         ":8000",
		Handler:      projectRouter.Handler(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("failed to start API server: %v", err)
	}
}
