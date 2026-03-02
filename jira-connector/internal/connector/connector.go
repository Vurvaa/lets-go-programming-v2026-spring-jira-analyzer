package connector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	URL "net/url"
)

type Project struct {
	ProjectId   string `json:"id"`
	ProjectName string `json:"name"`
	Issues      []json.RawMessage
}

const apiBase string = "/jira/rest/api/2/"

func GetBody(url string, expansion string) ([]byte, error) {
	resp, err := http.Get(url + apiBase + expansion)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("Response failed with status code: %d and\nbody: %s\n", resp.StatusCode, body)
	}
	if err != nil {
		return nil, err
	}
	return body, nil
}

func GetProjects(url string) ([]Project, error) {
	expansion := "project"
	body, err := GetBody(url, expansion)
	if err != nil {
		return nil, err
	}
	var projects []Project
	if typeErr := json.Unmarshal(body, &projects); typeErr != nil {
		return nil, fmt.Errorf("Type mismatch in API response: %w", typeErr)
	}
	return projects, nil
}

func GetIssues(url string, projects []Project, issueInOneRequest int) error {
	for i := 0; i < len(projects); i++ {
		startAt := 0
		for {
			escapedProjectName := `"` + URL.QueryEscape(projects[i].ProjectName) + `"`
			body, err := GetBody(url, fmt.Sprintf("search?jql=project=%s&expand=changelog&startAt=%d&maxResults=%d",
				escapedProjectName,
				startAt,
				issueInOneRequest,
			))
			if err != nil {
				return err
			}
			total := struct {
				Total  int               `json:"total"`
				Issues []json.RawMessage `json:"issues"`
			}{}
			if typeErr := json.Unmarshal(body, &total); typeErr != nil {
				return fmt.Errorf("Type mismatch in API response: %w", typeErr)
			}
			if len(projects[i].Issues) == 0 {
				projects[i].Issues = make([]json.RawMessage, 0, total.Total)
			}
			projects[i].Issues = append(projects[i].Issues, total.Issues...)
			startAt += len(total.Issues)
			if startAt >= total.Total {
				break
			}
		}
	}
	return nil
}
