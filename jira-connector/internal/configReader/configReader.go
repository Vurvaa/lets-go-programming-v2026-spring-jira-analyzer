package configReader

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"jira-connector/internal/apiServer/serverConfig"
)

type ConfigReader struct {
	path string
}

func NewConfigReader(path string) *ConfigReader {
	return &ConfigReader{path: path}
}

func (cr *ConfigReader) ReadServerConfig() (serverConfig.ServerConfig, error) {
	data, err := os.ReadFile(cr.path)
	if err != nil {
		return serverConfig.ServerConfig{}, fmt.Errorf("error reading config file %q: %w", cr.path, err)
	}

	var cfg serverConfig.ServerConfig
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return serverConfig.ServerConfig{}, fmt.Errorf("error parsing config file %q: %w", cr.path, err)
	}

	return cfg, nil
}
