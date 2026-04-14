package service

import (
	"encoding/json"
	"fmt"
	"server/internals/logger"
	"server/internals/model"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func (svc *ProjectService) compareOpenTaskTime(taskNum string, projectKeys []string) ([]byte, error) {
	if len(projectKeys) < 2 || len(projectKeys) > 3 {
		return nil, fmt.Errorf("count of projects for compare must be from 2 to 3")
	}

	graphs := make([]model.GraphData, 0, len(projectKeys))
	normalizedKeys := make([]string, 0, len(projectKeys))

	for _, key := range projectKeys {
		projectKey := strings.TrimSpace(key)
		if projectKey == "" {
			return nil, fmt.Errorf("project key is empty")
		}

		graph, err := svc.getGraph(taskNum, projectKey)
		if err != nil {
			logger.Instance.WithFields(logrus.Fields{
				"projectKey": projectKey,
				"error":      err,
			}).Error("get open task time graph for project")
			return nil, fmt.Errorf("get open task time graph for project %s: %w", projectKey, err)
		}

		if graph == nil {
			logger.Instance.WithField("projectKey", projectKey).
				Info("open task time graph for project was not found")
			return nil, fmt.Errorf("open task time graph for project %s was not found", projectKey)
		}

		graphs = append(graphs, *graph)
		normalizedKeys = append(normalizedKeys, projectKey)
	}

	categories := buildOpenTaskTimeCategories()

	result := model.CompareGraphData{
		Data:  categories,
		Count: make(map[string][]int, len(categories)),
	}

	for _, category := range categories {
		result.Count[category] = make([]int, len(graphs))

		for i, graph := range graphs {
			if value, exists := graph.Count[category]; exists {
				result.Count[category][i] = value
			}
		}
	}

	response, err := json.Marshal(result)
	if err != nil {
		logger.Instance.WithError(err).Error("marshal compare open task time response")
		return nil, fmt.Errorf("marshal compare open task time response: %w", err)
	}

	return response, nil
}

func (svc *ProjectService) compareTaskPriorityCount(taskNum string, projectKeys []string) ([]byte, error) {
	graphs := make([]model.GraphData, 0, len(projectKeys))
	categories := make([]string, 0)
	seen := make(map[string]struct{})

	for _, key := range projectKeys {
		key = strings.TrimSpace(key)
		graph, err := svc.getGraph(taskNum, key)
		if err != nil {
			logger.Instance.WithFields(logrus.Fields{
				"projectKey": key,
				"error":      err,
			}).Error("get task priority graph for project")
			return nil, fmt.Errorf("get task priority graph for project %s: %w", key, err)
		}
		if graph == nil {
			logger.Instance.WithField("projectKey", key).Info("task priority graph for project was not found")
			return nil, fmt.Errorf("task priority graph for project %s was not found", key)
		}

		graphs = append(graphs, *graph)

		for _, category := range graph.Categories {
			if _, exists := seen[category]; !exists {
				seen[category] = struct{}{}
				categories = append(categories, category)
			}
		}
	}

	result := model.CompareGraphData{
		Data:  categories,
		Count: make(map[string][]int),
	}

	for _, category := range categories {
		result.Count[category] = make([]int, len(projectKeys))

		for i, graph := range graphs {
			if value, exists := graph.Count[category]; exists {
				result.Count[category][i] = value
			}
		}
	}

	response, err := json.Marshal(result)
	if err != nil {
		logger.Instance.WithError(err).Error("marshal compare task priority count response")
		return nil, fmt.Errorf("marshal compare task priority count response: %w", err)
	}

	return response, nil
}

func (svc *ProjectService) buildOpenTaskTime(projectID string) error {
	rows, err := svc.projectRepository.GetIssueOpenTimesByProjectID(projectID)
	if err != nil {
		logger.Instance.WithFields(logrus.Fields{
			"projectID": projectID,
			"error":     err,
		}).Error("get issue open task times for project")
		return fmt.Errorf("get issue open times for project %s: %w", projectID, err)
	}

	categories := buildOpenTaskTimeCategories()
	count := make(map[string]int, len(categories))
	for _, category := range categories {
		count[category] = 0
	}

	closedTasksCount := 0

	for _, row := range rows {
		if !row.ClosedTime.Valid {
			continue
		}

		if row.ClosedTime.Time.Before(row.CreatedTime) {
			continue
		}

		openDuration := row.ClosedTime.Time.Sub(row.CreatedTime)

		category := getOpenTaskTimeCategory(openDuration)
		count[category]++
		closedTasksCount++
	}

	if closedTasksCount == 0 {
		logger.Instance.WithField("projectID", projectID).
			Info("no closed tasks for open task time histogram")
		return fmt.Errorf("it is not possible to build a histogram: the project has no closed tasks")
	}

	graph := model.GraphData{
		Categories: categories,
		Count:      count,
	}

	data, err := json.Marshal(graph)
	if err != nil {
		logger.Instance.WithFields(logrus.Fields{
			"projectID": projectID,
			"error":     err,
		}).Error("marshal open task time graph for project")
		return fmt.Errorf("marshal open task time graph for project %s: %w", projectID, err)
	}

	if err := svc.projectRepository.SaveOpenTaskTime(projectID, data); err != nil {
		logger.Instance.WithFields(logrus.Fields{
			"projectID": projectID,
			"error":     err,
		}).Error("save open task time for project")
		return fmt.Errorf("save open task time graph for project %s: %w", projectID, err)
	}

	return nil
}

func (svc *ProjectService) buildTaskPriorityCount(projectID string) error {
	rows, err := svc.projectRepository.GetIssuePrioritiesByProjectID(projectID)
	if err != nil {
		logger.Instance.WithFields(logrus.Fields{
			"projectID": projectID,
			"error":     err,
		}).Error("get project priorities for project")
		return fmt.Errorf("get issue priorities for project %s: %w", projectID, err)
	}

	count := make(map[string]int)
	categories := make([]string, 0)

	for _, row := range rows {
		priority := row.Priority

		if _, exists := count[priority]; !exists {
			categories = append(categories, priority)
		}
		count[priority]++
	}

	graph := model.GraphData{
		Categories: categories,
		Count:      count,
	}

	data, err := json.Marshal(graph)
	if err != nil {
		logger.Instance.WithFields(logrus.Fields{
			"projectID": projectID,
			"error":     err,
		}).Error("marshal task priority count for project")
		return fmt.Errorf("marshal task priority count for project %s: %w", projectID, err)
	}

	if err := svc.projectRepository.SaveTaskPriorityCount(projectID, data); err != nil {
		logger.Instance.WithFields(logrus.Fields{
			"projectID": projectID,
			"error":     err,
		}).Error("save task priority count for project")
		return fmt.Errorf("save task priority count for project %s: %w", projectID, err)
	}

	return nil
}

func (svc *ProjectService) getGraph(taskNum, projectKey string) (*model.GraphData, error) {
	data, err := svc.GetGraphByProjectKey(taskNum, projectKey)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}

	var graph model.GraphData
	if err := json.Unmarshal(data, &graph); err != nil {
		logger.Instance.WithFields(logrus.Fields{
			"project_key": projectKey,
			"error":       err,
		}).Error("failed to unmarshal graph data for project")
		return nil, fmt.Errorf("unmarshal graph for project %s: %w", projectKey, err)
	}

	return &graph, nil
}

func buildOpenTaskTimeCategories() []string {
	categories := make([]string, 0)

	for i := 0; i <= 23; i++ {
		categories = append(categories, fmt.Sprintf("%dh", i))
	}

	for i := 1; i <= 30; i++ {
		categories = append(categories, fmt.Sprintf("%dday", i))
	}

	for i := 1; i <= 11; i++ {
		categories = append(categories, fmt.Sprintf("%dmonth", i))
	}

	for i := 1; i <= 7; i++ {
		categories = append(categories, fmt.Sprintf("%dyear", i))
	}

	categories = append(categories, "8+year")

	return categories
}

func getOpenTaskTimeCategory(openDuration time.Duration) string {
	hours := openDuration.Hours()

	switch {
	case hours < 24:
		hourBucket := int(hours)
		return fmt.Sprintf("%dh", hourBucket)

	case hours < 24*31:
		dayBucket := int(hours / 24)
		if dayBucket < 1 {
			dayBucket = 1
		}
		if dayBucket > 30 {
			dayBucket = 30
		}
		return fmt.Sprintf("%dday", dayBucket)

	case hours < 24*30*12:
		monthBucket := int(hours / (24 * 30))
		if monthBucket < 1 {
			monthBucket = 1
		}
		if monthBucket > 11 {
			monthBucket = 11
		}
		return fmt.Sprintf("%dmonth", monthBucket)

	case hours < 24*365*8:
		yearBucket := int(hours / (24 * 365))
		if yearBucket < 1 {
			yearBucket = 1
		}
		if yearBucket > 7 {
			yearBucket = 7
		}
		return fmt.Sprintf("%dyear", yearBucket)

	default:
		return "8+year"
	}
}
