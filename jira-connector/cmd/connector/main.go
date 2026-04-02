package main

import (
	"jira-connector/internal/apiServer"
	"jira-connector/internal/connector"
)

func main() {
	connector.InitParameters("config.yaml")
	apiServer.NewServer().Start()
}
