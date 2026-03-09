package main

import (
	"context"
	"jira-connector/internal/connector"
	"jira-connector/internal/dataTransformer"
	"jira-connector/internal/pusher"
	"log"
)

func main() {
	url := "https://issues.apache.org"

	projets, err := connector.GetProjects(url)
	if err != nil {
		log.Fatal(err)
	}

	if err = connector.GetIssues(url, projets, 1000); err != nil {
		log.Fatal(err)
	}

	parsedIssues, err := dataTransformer.ParseIssues(projets)
	if err != nil {
		log.Fatal(err)
	}

	connectionString := "postgres://pguser:pgpwd@jira_postgres:5432/jira_db?sslmode=disable"
	store, err := pusher.NewStorage(connectionString)
	if err != nil {
		log.Fatal("Failed to connect to store:", err)
	}
	defer store.Close()

	ctx := context.Background()
	if err := store.SaveAll(ctx, parsedIssues, projets); err != nil {
		log.Fatal("Failed to save data:", err)
	}

	log.Println("Successfully saved data")
}
