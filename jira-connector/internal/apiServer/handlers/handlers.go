package handlers

import (
	"jira-connector/internal/apiServer/projectModels"
	"jira-connector/internal/connector"
	"math"
	"net/http"
	"strconv"
	"strings"
)

func GetProjectResponse(page, limit int, projects []projectModels.Project) projectModels.ProjectResponse {
	projectsCount := len(projects)
	startIndex := (page - 1) * limit
	endIndex := startIndex + limit
	if endIndex >= len(projects) {
		endIndex = len(projects)
	}

	return projectModels.ProjectResponse{
		Projects: projects[startIndex:endIndex],
		PageInfo: projectModels.PageInfo{
			CurrentPage:   page,
			PageCount:     int(math.Ceil(float64(projectsCount) / float64(limit))),
			ProjectsCount: projectsCount,
		},
	}
}

func HandleProjects(url, search string) ([]projectModels.Project, error) {
	projects, err := connector.GetProjects(url)
	if err != nil {
		return nil, err
	}

	var responseProjects []projectModels.Project

	for _, project := range projects {
		isCorrectName := strings.Contains(strings.ToLower(project.ProjectName), strings.ToLower(search))
		if isCorrectName {
			responseProject := projectModels.Project{
				ProjectId:   project.ProjectId,
				ProjectName: project.ProjectName,
				Existence:   false}

			responseProjects = append(responseProjects, responseProject)
		}
	}

	return responseProjects, nil
}

func ParseProjectParameters(request *http.Request) (int, int, string) {
	limit := 20
	page := 1
	search := ""

	limitPrm := request.URL.Query().Get("limit")
	if len(limitPrm) != 0 {
		limit, _ = strconv.Atoi(limitPrm)
	}

	pagePrm := request.URL.Query().Get("page")
	if len(pagePrm) != 0 {
		page, _ = strconv.Atoi(pagePrm)
	}

	searchPrm := request.URL.Query().Get("search")
	if len(searchPrm) != 0 {
		search = searchPrm
	}

	return limit, page, search
}
