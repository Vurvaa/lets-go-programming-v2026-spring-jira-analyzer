package dataTransformer

import (
	"encoding/json"
	"fmt"
	"jira-connector/internal/connector"
)

type StatusChanges struct {
	AuthorName string
	ChangeTime string
	FromStatus string
	ToStatus   string
}

type Issue struct {
	IssueID string `json:"id"`
	Fields  struct {
		Project struct {
			ProjectID string `json:"id"`
		} `json:"project"`
		Creator struct {
			AuthorName string `json:"name"`
		} `json:"creator"`
		Assignee struct {
			AssigneeName string `json:"name"`
			Key          string `json:"key"`
		} `json:"assignee"`
		Summary     string `json:"summary"`
		Description string `json:"description"`
		IssueType   struct {
			Type string `json:"name"`
		} `json:"issuetype"`
		Priority struct {
			Name string `json:"name"`
		} `json:"priority"`
		Status struct {
			Name string `json:"name"`
		} `json:"status"`
		CreatedTime string `json:"created"`
		UpdatedTime string `json:"updated"`
		TimeSpent   int    `json:"timespent"`
	} `json:"fields"`
	StatusChanges []StatusChanges
}

func parseStatusChanges(jsonData []byte) ([]StatusChanges, error) {
	var data struct {
		Changelog struct {
			Histories []struct {
				Author struct {
					Name string `json:"name"`
				} `json:"author"`
				Created string `json:"created"`
				Items   []struct {
					Field      string `json:"field"`
					FromString string `json:"fromString"`
					ToString   string `json:"toString"`
				} `json:"items"`
			} `json:"histories"`
		} `json:"changelog"`
	}
	if typeErr := json.Unmarshal(jsonData, &data); typeErr != nil {
		return nil, fmt.Errorf("Type mismatch in API response: %w", typeErr)
	}
	var changes []StatusChanges
	for _, h := range data.Changelog.Histories {
		for _, item := range h.Items {
			if item.Field == "status" {
				changes = append(changes, StatusChanges{
					AuthorName: h.Author.Name,
					ChangeTime: h.Created,
					FromStatus: item.FromString,
					ToStatus:   item.ToString,
				})
			}
		}
	}
	return changes, nil
}

func ParseIssuesOfProject(project *connector.Project) ([]Issue, error) {
	var allIssues []Issue
	for _, rawIssue := range project.Issues {
		var issue Issue
		if typeErr := json.Unmarshal(rawIssue, &issue); typeErr != nil {
			return nil, fmt.Errorf("Type mismatch in API response: %w", typeErr)
		}
		changes, err := parseStatusChanges(rawIssue)
		if err != nil {
			return nil, err
		}
		issue.StatusChanges = changes
		allIssues = append(allIssues, issue)
	}
	return allIssues, nil
}
