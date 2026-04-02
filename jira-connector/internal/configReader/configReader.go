package configReader

import (
	"fmt"
	"jira-connector/internal/apiServer/models"
	"os"

	"gopkg.in/yaml.v3"
)

type ConfigReader struct {
	path string
}

func NewConfigReader(path string) *ConfigReader {
	return &ConfigReader{path: path}
}

func (cr *ConfigReader) ReadServerConfig() (models.ServerConfig, error) {
	data, err := os.ReadFile(cr.path)
	if err != nil {
		return models.ServerConfig{}, fmt.Errorf("error reading config file %q: %w", cr.path, err)
	}

	var cfg models.ServerConfig
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return models.ServerConfig{}, fmt.Errorf("error parsing config file %q: %w", cr.path, err)
	}

	return cfg, nil
}

func (cr *ConfigReader) ReadConfigDB() (models.DBConfig, error) {
	data, err := os.ReadFile(cr.path)
	if err != nil {
		return models.DBConfig{}, fmt.Errorf("error reading config file %q: %w", cr.path, err)
	}

	var cfg models.DBConfig
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return models.DBConfig{}, fmt.Errorf("error parsing config file %q: %w", cr.path, err)
	}

	return cfg, nil
}
