package apiServer

import (
	"encoding/json"
	"fmt"
	"jira-connector/internal/apiServer/projectModels"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"jira-connector/internal/apiServer/serverConfig"
	"jira-connector/internal/configReader"
	"jira-connector/internal/connector"
)

type Server struct {
	configReader *configReader.ConfigReader
	config       *serverConfig.ServerConfig
}

func NewServer() *Server {
	reader := configReader.NewConfigReader("config.yaml")
	cfg, err := reader.ReadServerConfig()
	if err != nil {
		log.Fatal(err)
	}

	return &Server{
		configReader: reader,
		config:       serverConfig.NewServerConfig(cfg.ConnectorHost, cfg.ConnectorPort),
	}
}

func (server *Server) routes() {
	http.HandleFunc("/api/v1/connector/projects", server.projects)
	http.HandleFunc("/api/v1/connector/updateProject", server.updateProject)
}

func (server *Server) updateProject(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		http.Error(writer, "incorrect http method for /updateProject", http.StatusBadRequest)

		return
	}

	projectKey := request.URL.Query().Get("project")
	if projectKey == "" {
		http.Error(writer, "project name was not passed to /updateProject", http.StatusBadRequest)

		return
	}

	/*	issues, err := connector.GetProjectIssues(projectKey)
		if err != nil {
			http.Error(writer, fmt.Sprintf("error while downloading issues for project %q: %v", projectKey, err), http.StatusBadRequest)

			return
		}

		transformedIssues := server.dataTransformer.TransformIssues(issues)

		databasePusher.PushIssues(transformedIssues)*/
}

func (server *Server) projects(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		http.Error(writer, "incorrect http method for /projects", http.StatusBadRequest)

		return
	}

	limit, page, search := parseProjectParameters(request)

	projects, err := handleProjects(search)
	if err != nil {
		http.Error(writer, fmt.Sprintf("error while downloading list of projects: %v", err), http.StatusBadRequest)

		return
	}

	responseProjects := getProjectResponse(page, limit, projects)

	writer.Header().Set("Content-Type", "application/json")
	response, err := json.Marshal(responseProjects)
	if err != nil {
		http.Error(writer, fmt.Sprintf("error while marshalling response: %v", err), http.StatusInternalServerError)
	}

	_, _ = writer.Write(response)
}

func getProjectResponse(page, limit int, projects []projectModels.Project) projectModels.ProjectResponse {
	projectsCount := len(projects)
	startIndex := (page - 1) * limit
	endIndex := startIndex + limit

	return projectModels.ProjectResponse{
		Projects: projects[startIndex:endIndex],
		PageInfo: projectModels.PageInfo{
			CurrentPage:   page,
			PageCount:     int(math.Ceil(float64(projectsCount) / float64(limit))),
			ProjectsCount: projectsCount,
		},
	}
}

func handleProjects(search string) ([]projectModels.Project, error) {
	projects, err := connector.GetProjects("https://issues.apache.org")
	if err != nil {
		return nil, err
	}

	var responseProjects []projectModels.Project

	for _, project := range projects {
		isCorrectName := strings.Contains(strings.ToLower(project.ProjectName), strings.ToLower(search))
		if isCorrectName {
			responseProject := projectModels.Project{ProjectId: project.ProjectId, ProjectName: project.ProjectName}
			responseProjects = append(responseProjects, responseProject)
		}
	}

	return responseProjects, nil
}

func parseProjectParameters(request *http.Request) (int, int, string) {
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

func (server *Server) Start() {
	server.routes()

	addr := fmt.Sprintf("%s:%d", server.config.ConnectorHost, server.config.ConnectorPort)
	fmt.Println("listening on", addr)

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err)
	}
}
