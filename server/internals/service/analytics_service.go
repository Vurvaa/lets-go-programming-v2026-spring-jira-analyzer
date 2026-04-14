package service

import (
	"fmt"
	"server/internals/logger"
	"strconv"
)

func (svc *ProjectService) GetGraphByProjectKey(taskNum, projectKey string) ([]byte, error) {
	num, err := strconv.Atoi(taskNum)
	if err != nil {
		logger.Instance.WithError(err).Error("Error while converting taskNum to int")
		return nil, err
	}

	projectID, err := svc.projectRepository.GetProjectIDByKey(projectKey)
	if err != nil {
		logger.Instance.WithError(err).Error("Error while getting projectID by key")
		return nil, err
	}

	var body []byte = nil

	switch num {
	case 1:
		body, err = svc.projectRepository.GetOpenTaskTime(projectID)
	case 5:
		body, err = svc.projectRepository.GetTaskPriorityCount(projectID)
	default:
		logger.Instance.WithField("num", num).Info("Unknown task number")
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
		logger.Instance.WithError(err).Error("Error while converting taskNum to int")
		return err
	}

	projectID, err := svc.projectRepository.GetProjectIDByKey(projectKey)
	if err != nil {
		logger.Instance.WithError(err).Error("Error while getting projectID by key")
		return err
	}

	switch num {
	case 1:
		err = svc.buildOpenTaskTime(projectID)
	case 5:
		err = svc.buildTaskPriorityCount(projectID)
	default:
		logger.Instance.WithField("num", num).Info("Unknown task number")
		err = fmt.Errorf("unknown task number: %v", num)
	}

	return err
}

func (svc *ProjectService) CompareProjects(taskNum string, projectKey []string) ([]byte, error) {
	if len(projectKey) < 2 || len(projectKey) > 3 {
		logger.Instance.Error("count of projects for compare must be from 2 to 3")
		return nil, fmt.Errorf("count of projects for compare must be from 2 to 3")
	}

	num, err := strconv.Atoi(taskNum)
	if err != nil {
		logger.Instance.WithError(err).Error("Error while converting taskNum to int")
		return nil, err
	}

	switch num {
	case 1:
		return svc.compareOpenTaskTime(taskNum, projectKey)
	case 5:
		return svc.compareTaskPriorityCount(taskNum, projectKey)
	default:
		logger.Instance.WithField("num", num).Info("Unknown task number")
		return nil, fmt.Errorf("unknown task number: %v", num)
	}
}
