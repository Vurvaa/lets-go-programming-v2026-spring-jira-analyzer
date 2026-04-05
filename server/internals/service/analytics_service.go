package service

import (
	"encoding/json"
	"fmt"
	"log"
	"server/internals/model"
	"strconv"
	"strings"
	"time"
)

func (svc *ProjectService) GetGraphByProjectKey(taskNum, projectKey string) ([]byte, error) {
	num, err := strconv.Atoi(taskNum)
	if err != nil {
		log.Printf("Error while converting taskNum to int: %v", err)
		return nil, err
	}

	projectID, err := svc.projectRepository.GetProjectIDByKey(projectKey)
	if err != nil {
		log.Printf("Error while getting projectID by key: %v", err)
		return nil, err
	}

	var body []byte = nil

	switch num {
	case 1:
		body, err = svc.projectRepository.GetOpenTaskTime(projectID)
	case 5:
		body, err = svc.projectRepository.GetTaskPriorityCount(projectID)
	default:
		log.Printf("Unknown task number: %v", num)
		err = fmt.Errorf("unknown task number: %v", num)
	}

	return body, err
}

func (svc *ProjectService) DeleteAllGraphByProjectKey(projectKey string) error {
	projectID, err := svc.projectRepository.GetProjectIDByKey(projectKey)
	if err != nil {
		return err
	}

	return svc.projectRepository.DeleteProjectAnalytics(projectID)
}

func (svc *ProjectService) HasAnyAnalyticsByProjectKey(projectKey string) (bool, error) {
	projectID, err := svc.projectRepository.GetProjectIDByKey(projectKey)
	if err != nil {
		return false, err
	}

	return svc.projectRepository.HasAnyAnalytics(projectID)
}

func (svc *ProjectService) MakeGraphByProjectKey(taskNum, projectKey string) error {
	num, err := strconv.Atoi(taskNum)
	if err != nil {
		log.Printf("Error while converting taskNum to int: %v", err)
		return err
	}

	projectID, err := svc.projectRepository.GetProjectIDByKey(projectKey)
	if err != nil {
		log.Printf("Error while getting projectID by key: %v", err)
		return err
	}

	switch num {
	case 1:
		err = svc.buildOpenTaskTime(projectID)
	case 5:
		err = svc.buildTaskPriorityCount(projectID)
	default:
		log.Printf("Unknown task number: %v", num)
		err = fmt.Errorf("unknown task number: %v", num)
	}

	return err
}

func (svc *ProjectService) CompareProjects(taskNum string, projectKey []string) ([]byte, error) {
	if len(projectKey) < 2 || len(projectKey) > 3 {
		return nil, fmt.Errorf("count of projects for compare must be from 2 to 3")
	}

	num, err := strconv.Atoi(taskNum)
	if err != nil {
		log.Printf("Error while converting taskNum to int: %v", err)
		return nil, err
	}

	switch num {
	case 1:
		return svc.compareOpenTaskTime(taskNum, projectKey)
	case 5:
		return svc.compareTaskPriorityCount(taskNum, projectKey)
	default:
		log.Printf("Unknown task number: %v", num)
		return nil, fmt.Errorf("unknown task number: %v", num)
	}
}

func (svc *ProjectService) compareOpenTaskTime(taskNum string, projectKeys []string) ([]byte, error) {
	graphs := make([]model.GraphData, 0, len(projectKeys))

	for _, key := range projectKeys {
		projectKey := strings.TrimSpace(key)
		graph, err := svc.getGraph(taskNum, projectKey)
		if err != nil {
			return nil, fmt.Errorf("get open task time graph for project %s: %w", projectKey, err)
		}
		if graph == nil {
			return nil, fmt.Errorf("open task time graph for project %s was not found", projectKey)
		}

		graphs = append(graphs, *graph)
	}

	categories := []string{"0-1day", "1-3days", "3-7days", "7+days"}

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
			return nil, fmt.Errorf("get task priority graph for project %s: %w", key, err)
		}
		if graph == nil {
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
		return nil, fmt.Errorf("marshal compare task priority count response: %w", err)
	}

	return response, nil
}

func (svc *ProjectService) buildOpenTaskTime(projectID string) error {
	rows, err := svc.projectRepository.GetIssueOpenTimesByProjectID(projectID)
	if err != nil {
		return fmt.Errorf("get issue open times for project %s: %w", projectID, err)
	}

	graph := model.GraphData{
		Categories: []string{"0-1day", "1-3days", "3-7days", "7+days"},
		Count: map[string]int{
			"0-1day":  0,
			"1-3days": 0,
			"3-7days": 0,
			"7+days":  0,
		},
	}

	now := time.Now()

	for _, row := range rows {
		var openDuration time.Duration

		if row.ClosedTime.Valid {
			if row.ClosedTime.Time.Before(row.CreatedTime) {
				continue
			}
			openDuration = row.ClosedTime.Time.Sub(row.CreatedTime)
		} else {
			openDuration = now.Sub(row.CreatedTime)
		}

		switch {
		case openDuration <= 24*time.Hour:
			graph.Count["0-1day"]++
		case openDuration <= 3*24*time.Hour:
			graph.Count["1-3days"]++
		case openDuration <= 7*24*time.Hour:
			graph.Count["3-7days"]++
		default:
			graph.Count["7+days"]++
		}
	}

	data, err := json.Marshal(graph)
	if err != nil {
		return fmt.Errorf("marshal open task time graph for project %s: %w", projectID, err)
	}

	if err := svc.projectRepository.SaveOpenTaskTime(projectID, data); err != nil {
		return fmt.Errorf("save open task time graph for project %s: %w", projectID, err)
	}

	return nil
}

func (svc *ProjectService) buildTaskPriorityCount(projectID string) error {
	rows, err := svc.projectRepository.GetIssuePrioritiesByProjectID(projectID)
	if err != nil {
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
		return fmt.Errorf("marshal task priority count for project %s: %w", projectID, err)
	}

	if err := svc.projectRepository.SaveTaskPriorityCount(projectID, data); err != nil {
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
		return nil, fmt.Errorf("unmarshal graph for project %s: %w", projectKey, err)
	}

	return &graph, nil
}
