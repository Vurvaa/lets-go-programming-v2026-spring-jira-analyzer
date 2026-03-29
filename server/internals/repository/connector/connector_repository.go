package connector

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"server/internals/config"
)

type Repository struct {
	connectorURL string
}

func NewConnectorRepository(configName string) *Repository {
	connectorConfig := config.LoadConnectorConfig(configName)
	if connectorConfig == nil {
		log.Printf("Unable to load connector config from file %s", configName)
		return nil
	}

	return &Repository{
		connectorURL: fmt.Sprintf(
			"http://%s:%d",
			connectorConfig.ConnectorHost,
			connectorConfig.ConnectorPort,
		),
	}
}

func (cntr *Repository) GetAllProjects(rawQuery string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/v1/projects", cntr.connectorURL)
	if rawQuery != "" {
		url += "?" + rawQuery
	}

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error while requesting connector /projects: %v", err)
		return nil, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Printf("Unable to Close() connector response body")
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error while reading connector /projects response: %v", err)
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("Connector /projects returned status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("connector returned status %d", resp.StatusCode)
	}

	return body, nil
}

func (cntr *Repository) UpdateProject(rawQuery string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/v1/updateProject", cntr.connectorURL)
	if rawQuery != "" {
		url += "?" + rawQuery
	}

	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		log.Printf("Error while requesting connector /updateProject: %v", err)
		return nil, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Printf("Unable to Close() connector response body")
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error while reading connector /updateProject response: %v", err)
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("Connector /updateProject returned status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("connector returned status %d", resp.StatusCode)
	}

	return body, nil
}
