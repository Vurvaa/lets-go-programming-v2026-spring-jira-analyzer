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
		err = connector.GetIssues(url, &project, 200, 30)
		if err != nil {
			log.Fatal(err)
		}
	}
}
