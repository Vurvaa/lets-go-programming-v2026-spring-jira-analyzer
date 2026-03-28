package main

import (
	"log"
	"time"

	repository "server/internals/repository/connector"
	"server/internals/repository/postgres"
	"server/internals/service"
)

func main() {
	const configName = "config.yaml"

	time.Sleep(45 * time.Second)

	db := postgres.NewDB(configName)
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to close db: %v", err)
		}
	}()

	projectRepo := postgres.NewProjectRepository(db)
	connector := repository.NewConnectorRepository(configName)
	projectService := service.NewProjectService(connector, projectRepo)

	projects, err := projectService.GetAllProjectsFromDB()
	if err != nil {
		log.Fatalf("failed to get projects: %v", err)
	}

	log.Printf("projects: %+v", projects)
}
