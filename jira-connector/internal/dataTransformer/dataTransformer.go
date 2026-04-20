package dataTransformer

import (
	"encoding/json"
	"fmt"
	"jira-connector/internal/connector"
	"jira-connector/internal/logger"

	"github.com/sirupsen/logrus"
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
		ClosedTime  string `json:"resolutiondate"`
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
		logger.Instance.WithError(typeErr).Error("Failed to unmarshal changelog")
		return nil, fmt.Errorf("type mismatch in API response: %w", typeErr)
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
	logger.Instance.WithFields(logrus.Fields{
		"project_key": project.ProjectKey,
		"raw_issues":  len(project.Issues),
	}).Info("Starting data transformation")

	var allIssues []Issue
	for i, rawIssue := range project.Issues {
		var issue Issue
		if typeErr := json.Unmarshal(rawIssue, &issue); typeErr != nil {
			logger.Instance.WithError(typeErr).WithField("issue_index", i).Error("Failed to unmarshal individual issue")
			return nil, fmt.Errorf("type mismatch in API response: %w", typeErr)
		}
		changes, err := parseStatusChanges(rawIssue)
		if err != nil {
			logger.Instance.WithError(err).WithField("issue_id", issue.IssueID).Warn("Could not parse status changes for issue")
		}
		issue.StatusChanges = changes
		allIssues = append(allIssues, issue)
	}

	logger.Instance.WithField("count", len(allIssues)).Info("Data transformation completed")
	return allIssues, nil
}
