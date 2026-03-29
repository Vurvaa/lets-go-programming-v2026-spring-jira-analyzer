package service

import (
	"encoding/json"
	"log"
	"server/internals/model"
	connectorRepo "server/internals/repository/connector"
	postgresRepo "server/internals/repository/postgres"
)

type ProjectService struct {
	connectorRepository *connectorRepo.Repository
	projectRepository   *postgresRepo.ProjectRepository
}

func NewProjectService(
	connectorRepository *connectorRepo.Repository,
	projectRepository *postgresRepo.ProjectRepository,
) *ProjectService {
	return &ProjectService{
		connectorRepository: connectorRepository,
		projectRepository:   projectRepository,
	}
}

func (s *ProjectService) GetAllProjectsFromDB() ([]model.Project, error) {
	return s.projectRepository.GetAllProjects()
}

func (s *ProjectService) GetAllProjectsFromRepository(rawQuery string) ([]byte, error) {
	body, err := s.connectorRepository.GetAllProjects(rawQuery)
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
		exists, err := s.projectRepository.ProjectExistsByID(projectResponse.Projects[i].ProjectID)
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

func (s *ProjectService) DeleteProject(id string) error {
	return s.projectRepository.Delete(id)
}

func (s *ProjectService) GetProjectStatsByID(projectID string) (model.ProjectStats, error) {
	stats, err := s.projectRepository.GetProjectStatsByID(projectID)
	if err != nil {
		return stats, err
	}

	reopenedIssuesCount, err := s.projectRepository.GetReopenedIssuesCount(projectID)
	if err != nil {
		return stats, err
	}
	stats.ReopenedIssuesCount = reopenedIssuesCount

	averageIssuesCount, err := s.projectRepository.GetAverageIssuesCount(projectID)
	if err != nil {
		return stats, err
	}
	stats.AverageIssuesCount = averageIssuesCount

	averageTime, err := s.projectRepository.GetAverageTime(projectID)
	if err != nil {
		return stats, err
	}
	stats.AverageTime = averageTime

	stats.Key = stats.ProjectID

	return stats, nil
}

func (s *ProjectService) UpdateProject(rawQuery string) ([]byte, error) {
	return s.connectorRepository.UpdateProject(rawQuery)
}
