package main

import (
	"log"

	"server/internals/repository/postgres"
	"server/internals/service"
)

func main() {
	db := postgres.NewDB()
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to close db: %v", err)
		}
	}()

	projectRepo := postgres.NewProjectRepository(db)
	projectService := service.NewProjectService(projectRepo)

	projects, err := projectService.GetAllProjects(0, 10)
	if err != nil {
		log.Fatalf("failed to get projects: %v", err)
	}

	println("projects: ", projects)

	log.Printf("projects: %+v", projects)
}
