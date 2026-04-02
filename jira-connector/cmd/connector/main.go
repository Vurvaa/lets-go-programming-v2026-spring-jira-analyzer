package main

import (
	"fmt"
	"jira-connector/internal/apiServer"
	"jira-connector/internal/configReader"
	"jira-connector/internal/connector"
	"jira-connector/internal/pusher"
	"log"
	"time"
)

func main() {
	const configName = "config.yaml"

	connector.InitParameters(configName)

	reader := configReader.NewConfigReader(configName)
	cfg, err := reader.ReadConfigDB()
	if err != nil {
		log.Fatal(err)
		return
	}

	connectionStr := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.UserDB,
		cfg.PasswordDB,
		cfg.HostDB,
		cfg.PortDB,
		cfg.NameDB,
	)

	var store *pusher.Storage
	for i := 0; i < 15; i++ {
		store, err = pusher.NewStorage(connectionStr)
		if err == nil {
			log.Println("Successfully connected to database")
			break
		}
		log.Printf("Database not ready (attempt %d/15), waiting...", i+1)
		time.Sleep(2 * time.Second)
	}

	srvConfig, err := reader.ReadServerConfig()
	if err != nil {
		log.Fatal(err)
		return
	}

	apiServer.NewServer(srvConfig, store).Start()
}
