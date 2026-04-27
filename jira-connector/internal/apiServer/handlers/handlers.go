package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"jira-connector/internal/apiServer/models"
	"jira-connector/internal/connector"
)

const (
	maxLimit     = 100
	defaultLimit = 20
	defaultPage  = 1
)

func GetProjectResponse(page, limit int, projects []models.Project) models.ProjectResponse {
	projectsCount := len(projects)

	pageCount := 0
	if projectsCount > 0 {
		pageCount = (projectsCount + limit - 1) / limit
	}

	startIndex := (page - 1) * limit
	if startIndex > projectsCount {
		startIndex = projectsCount
	}

	endIndex := startIndex + limit
	if endIndex > projectsCount {
		endIndex = projectsCount
	}

	return models.ProjectResponse{
		Projects: projects[startIndex:endIndex],
		PageInfo: models.PageInfo{
			CurrentPage:   page,
			PageCount:     pageCount,
			ProjectsCount: projectsCount,
		},
	}
}

func HandleProjects(url, search string) ([]models.Project, error) {
	projects, err := connector.GetProjects(url)
	if err != nil {
		return nil, err
	}

	responseProjects := make([]models.Project, 0)

	for _, project := range projects {
		isCorrectName := strings.Contains(strings.ToLower(project.ProjectName), strings.ToLower(search))
		if isCorrectName {
			responseProject := models.Project{
				ProjectId:   project.ProjectId,
				ProjectKey:  project.ProjectKey,
				ProjectName: project.ProjectName,
				ProjectUrl:  project.ProjectSelf,
				Existence:   false}

			responseProjects = append(responseProjects, responseProject)
		}
	}

	return responseProjects, nil
}

func ParseLimit(request *http.Request) int {
	limitPrm := request.URL.Query().Get("limit")
	if limitPrm == "" {
		return defaultLimit
	}

	limit, err := strconv.Atoi(limitPrm)
	if err != nil || limit <= 0 {
		return defaultLimit
	}

	if limit > maxLimit {
		return maxLimit
	}

	return limit
}

func ParsePage(request *http.Request) int {
	pagePrm := request.URL.Query().Get("page")
	if pagePrm == "" {
		return defaultPage
	}

	page, err := strconv.Atoi(pagePrm)
	if err != nil || page <= 0 {
		return defaultPage
	}

	return page
}

func ParseSearch(request *http.Request) string {
	return request.URL.Query().Get("search")
}
