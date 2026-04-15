package service

import (
	"encoding/json"
	"log"
	"math"
	"server/internals/model"
	connectorRepo "server/internals/repository/connector"
	postgresRepo "server/internals/repository/postgres"
)

type ProjectService struct {
	connectorRepository *connectorRepo.Repository
	projectRepository   *postgresRepo.DBRepository
}

func NewProjectService(
	connectorRepository *connectorRepo.Repository,
	projectRepository *postgresRepo.DBRepository,
) *ProjectService {
	return &ProjectService{
		connectorRepository: connectorRepository,
		projectRepository:   projectRepository,
	}
}

func (svc *ProjectService) GetAllProjectsFromDB() ([]model.Project, error) {
	return svc.projectRepository.GetAllProjects()
}

func (svc *ProjectService) GetAllProjectsFromRepository(rawQuery string) ([]byte, error) {
	body, err := svc.connectorRepository.GetAllProjects(rawQuery)
	if err != nil {
		return nil, err
	}

	var projectResponse model.ProjectResponse
	err = json.Unmarshal(body, &projectResponse)
	if err != nil {
		log.Printf("Error while unmarshalling connector response: %v", err)
		return nil, err
	}

	for i := range projectResponse.Projects {
		exists, err := svc.projectRepository.ProjectExistsByID(projectResponse.Projects[i].ProjectID)
		if err != nil {
			return nil, err
		}
		projectResponse.Projects[i].Existence = exists
	}

	updatedBody, err := json.Marshal(projectResponse)
	if err != nil {
		log.Printf("Error while marshalling updated project response: %v", err)
		return nil, err
	}

	return updatedBody, nil
}

func (svc *ProjectService) DeleteProject(id string) error {
	return svc.projectRepository.DeleteProjectByID(id)
}

func (svc *ProjectService) GetProjectStatsByID(projectID string) (model.ProjectStats, error) {
	stats, err := svc.projectRepository.GetProjectStatsByID(projectID)
	if err != nil {
		return stats, err
	}

	reopenedIssuesCount, err := svc.projectRepository.GetReopenedIssuesCount(projectID)
	if err != nil {
		return stats, err
	}
	stats.ReopenedIssuesCount = reopenedIssuesCount

	averageIssuesCount, err := svc.projectRepository.GetAverageIssuesCount(projectID)
	if err != nil {
		return stats, err
	}
	stats.AverageIssuesCount = averageIssuesCount

	averageTime, err := svc.projectRepository.GetAverageTime(projectID)
	if err != nil {
		return stats, err
	}
	stats.AverageTime = math.Round(averageTime*100) / 100

	return stats, nil
}

func (svc *ProjectService) UpdateProject(rawQuery string) ([]byte, error) {
	return svc.connectorRepository.UpdateProject(rawQuery)
}
