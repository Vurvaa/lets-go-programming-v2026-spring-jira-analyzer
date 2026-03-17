package service

import (
	"server/internals/model"
	"server/internals/repository/postgres"
)

type ProjectService struct {
	repo *postgres.ProjectRepository
}

func NewProjectService(repo *postgres.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) GetAllProjects(offset int, limit int) ([]model.Project, error) {
	return s.repo.GetAllProjects(offset, limit)
}
