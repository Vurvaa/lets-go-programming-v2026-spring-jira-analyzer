package main

import (
	"jira-connector/internal/apiServer"
)

func main() {
	apiServer.NewServer().Start()
}
