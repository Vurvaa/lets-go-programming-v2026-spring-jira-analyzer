package main

import (
	"fmt"
	"jira-connector/internal/apiServer"
	"jira-connector/internal/config"
	"jira-connector/internal/connector"
	"jira-connector/internal/logger"
	"jira-connector/internal/pusher"
	"time"
)

func main() {
	const configName = "config.yaml"
	cfg, err := config.LoadConfig(configName)
	if err != nil {
		logger.Instance.WithError(err).Fatal("Failed to read DB config")
	}
	connector.InitParameters(&cfg.Connector)
	connectionStr := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DB.UserDB,
		cfg.DB.PasswordDB,
		cfg.DB.HostDB,
		cfg.DB.PortDB,
		cfg.DB.NameDB,
	)
	var store *pusher.Storage
	for i := 0; i < 15; i++ {
		store, err = pusher.NewStorage(connectionStr)
		if err == nil {
			logger.Instance.Info("Successfully connected to database")
			break
		}
		logger.Instance.WithField("attempt", i+1).Warn("Database not ready, waiting...")
		time.Sleep(2 * time.Second)
	}
	if store == nil {
		logger.Instance.Fatal("Cannot start server: database connection failed")
	}
	apiServer.NewServer(cfg.Server, store).Start()
}
