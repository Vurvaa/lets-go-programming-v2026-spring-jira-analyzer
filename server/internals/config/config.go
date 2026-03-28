package config

import (
	"gopkg.in/yaml.v2"

	"log"
	"os"
)

func LoadDBConfig(filename string) *DBConfig {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("Error reading config file %s", filename)
		return nil
	}

	config := &DBConfig{}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		log.Printf("Error reading config file %s", filename)
		return nil
	}

	return config
}

func LoadConnectorConfig(filename string) *ConnectorConfig {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("Error reading config file %s", filename)
		return nil
	}

	config := &ConnectorConfig{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		log.Printf("Error reading config file %s", filename)
		return nil
	}

	return config
}
