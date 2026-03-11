package main

import (
	"context"
	"fmt"
	"jira-connector/internal/connector"
	"jira-connector/internal/dataTransformer"
	"jira-connector/internal/pusher"
	"log"
)

func main() {
	url := "https://issues.apache.org"
	projectPtr, err := connector.GetProject(url, "AAR")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(projectPtr)

	project := []connector.Project{*projectPtr}
	err = connector.GetIssues(url, project, 1000)
	if err != nil {
		log.Fatal(err)
	}

	issues, err := dataTransformer.ParseIssues(project)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(issues)

	projects, err := connector.GetProjects(url)
	if err != nil {
		log.Fatal(err)
	}

	if err = connector.GetIssues(url, projects, 1000); err != nil {
		log.Fatal(err)
	}

	parsedIssues, err := dataTransformer.ParseIssues(projects)
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
	if err := store.SaveAll(ctx, parsedIssues, projects); err != nil {
		log.Fatal("Failed to save data:", err)
	}

	log.Println("Successfully saved data")
}
