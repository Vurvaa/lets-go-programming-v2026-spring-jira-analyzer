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

func (serv *ProjectService) GetAllProjectsFromDB() ([]model.Project, error) {
	return serv.projectRepository.GetAllProjects()
}

func (serv *ProjectService) GetAllProjectsFromRepository(rawQuery string) ([]byte, error) {
	body, err := serv.connectorRepository.GetAllProjects(rawQuery)
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
		exists, err := serv.projectRepository.ProjectExistsByID(projectResponse.Projects[i].ProjectID)
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

func (serv *ProjectService) UpdateProject(rawQuery string) ([]byte, error) {
	return serv.connectorRepository.UpdateProject(rawQuery)
}
