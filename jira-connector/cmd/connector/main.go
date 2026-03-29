package main

import (
	"jira-connector/internal/connector"
	"log"
)

func main() {
	url := "https://issues.apache.org"
	projects, err := connector.GetProjects(url)
	if err != nil {
		log.Fatal(err)
	}
	for _, project := range projects {
		if err = connector.GetIssues(url, &project); err != nil {
			log.Fatal(err)
		}
	}
}
