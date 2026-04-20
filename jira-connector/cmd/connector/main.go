package main

import (
	"fmt"
	"jira-connector/internal/apiServer"
	"jira-connector/internal/configReader"
	"jira-connector/internal/connector"
	"jira-connector/internal/logger"
	"jira-connector/internal/pusher"
	"time"
)

func main() {
	const configName = "config.yaml"

	logger.InitLogger()
	connector.InitParameters(configName)

	reader := configReader.NewConfigReader(configName)
	cfg, err := reader.ReadConfigDB()
	if err != nil {
		logger.Instance.WithError(err).Fatal("Failed to read DB config")
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
			logger.Instance.Info("Successfully connected to database")
			break
		}
		logger.Instance.WithField("attempt", i+1).Warn("Database not ready, waiting...")
		time.Sleep(2 * time.Second)
	}

	srvConfig, err := reader.ReadServerConfig()
	if err != nil {
		logger.Instance.WithError(err).Fatal("Failed to read server config")
		return
	}

	logger.Instance.WithField("port", srvConfig.ConnectorPort).Info("Starting Jira Connector server")
	apiServer.NewServer(srvConfig, store).Start()
}
