package main

import (
	"jira-connector/internal/connector"
	"jira-connector/internal/dataTransformer"
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
	_, err = dataTransformer.ParseIssues(projets)
	if err != nil {
		log.Fatal(err)
	}
}
