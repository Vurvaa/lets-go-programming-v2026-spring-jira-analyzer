package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type FullConfig struct {
	DB        DBConfig        `yaml:"db"`
	Server    ServerConfig    `yaml:"server"`
	Connector ConnectorConfig `yaml:"connector"`
}

type DBConfig struct {
	HostDB     string `yaml:"db_host"`
	PortDB     int    `yaml:"db_port"`
	UserDB     string `yaml:"db_user"`
	PasswordDB string `yaml:"db_passwd"`
	NameDB     string `yaml:"db_name"`
}

type ServerConfig struct {
	Repository    string `yaml:"jiraUrl"`
	ConnectorPort uint   `yaml:"connector_port"`
	ConnectorHost string `yaml:"connector_host"`
}

type ConnectorConfig struct {
	MinTimeSleep      int64 `yaml:"minTimeSleep"`
	MaxTimeSleep      int64 `yaml:"maxTimeSleep"`
	Goroutines        int   `yaml:"threadCount"`
	IssueInOneRequest int   `yaml:"issueInOneRequest"`
}

func LoadConfig(path string) (*FullConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error of reading config - %s: %w", path, err)
	}
	var cfg FullConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("type mismatch in config file %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (cfg *FullConfig) Validate() error {
	if cfg.DB.HostDB == "" {
		return fmt.Errorf("db_host is empty")
	}
	if cfg.DB.PortDB == 0 {
		return fmt.Errorf("db_port is zero")
	}
	if cfg.DB.UserDB == "" {
		return fmt.Errorf("db_user is empty")
	}
	if cfg.DB.PasswordDB == "" {
		return fmt.Errorf("db_passwd is empty")
	}
	if cfg.DB.NameDB == "" {
		return fmt.Errorf("db_name is empty")
	}

	if cfg.Server.Repository == "" {
		return fmt.Errorf("jiraUrl is empty")
	}
	if cfg.Server.ConnectorPort == 0 {
		return fmt.Errorf("connector_port is zero")
	}
	if cfg.Server.ConnectorHost == "" {
		return fmt.Errorf("connector_host is empty")
	}

	if cfg.Connector.Goroutines < 1 {
		return fmt.Errorf("threadCount must be >=1, got %d", cfg.Connector.Goroutines)
	}
	if cfg.Connector.IssueInOneRequest < 50 || cfg.Connector.IssueInOneRequest > 1000 {
		return fmt.Errorf("issueInOneRequest must be [50; 1000], got %d", cfg.Connector.IssueInOneRequest)
	}
	return nil
}
