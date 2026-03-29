package main

import (
	"context"
	"jira-connector/internal/apiServer"
	"jira-connector/internal/connector"
	"jira-connector/internal/dataTransformer"
	"jira-connector/internal/pusher"
	"log"
	"time"
)

func main() {
	go runSyncProcess()

	log.Println("Starting Connector API Server...")
	apiServer.NewServer().Start()
}

func runSyncProcess() {
	time.Sleep(10 * time.Second)

	url := "https://issues.apache.org"
	log.Printf("Starting sync with %s", url)

	projects, err := connector.GetProjects(url)
	if err != nil {
		log.Printf("ERROR: Failed to get projects: %v", err)
		return
	}

	projects = projects[:5] // ВРЕМЕННОЕ ОГРАНИЧЕНИЕ ДЛЯ ТЕСТОВ
	for i := range projects {
		err = connector.GetIssues(url, &projects[i], 200, 30)
		if err != nil {
			log.Printf("WARNING: Failed to get issues for project %s: %v", projects[i].ProjectName, err)
		}
	}

	parsedIssues, err := dataTransformer.ParseIssues(projects)
	if err != nil {
		log.Printf("ERROR: Transformation failed: %v", err)
		return
	}

	connStr := "postgres://pguser:pgpwd@db:5432/jira_db?sslmode=disable"
	store, err := pusher.NewStorage(connStr)
	if err != nil {
		log.Printf("ERROR: Failed to connect to DB: %v", err)
		return
	}
	defer store.Close()

	if err := store.SaveAll(context.Background(), parsedIssues, projects); err != nil {
		log.Printf("ERROR: Failed to save data: %v", err)
		return
	}

	log.Println("SUCCESS: All data from Jira saved to Database")
}
