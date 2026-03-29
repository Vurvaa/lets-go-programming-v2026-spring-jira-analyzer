package service

import (
	"server/internals/model"
)

func (s *ProjectService) GetAllProjects() ([]model.Project, error) {
	return s.projectRepository.GetAllProjects()
}

func (s *ProjectService) GetProject(id string) (*model.Project, error) {
	return s.projectRepository.GetByID(id)
}

func (s *ProjectService) DeleteProject(id string) error {
	return s.projectRepository.Delete(id)
}
