package apiServer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"jira-connector/internal/apiServer/projectModels"
	"jira-connector/internal/apiServer/serverConfig"
	"jira-connector/internal/configReader"
	"jira-connector/internal/connector"
	"jira-connector/internal/dataTransformer"
	"jira-connector/internal/pusher"
)

type Server struct {
	configReader *configReader.ConfigReader
	config       *serverConfig.ServerConfig
	storage      *pusher.Storage
}

func NewServer() *Server {
	reader := configReader.NewConfigReader("config.yaml")
	cfg, err := reader.ReadServerConfig()
	if err != nil {
		log.Fatal(err)
	}

	connectionString := "postgres://pguser:pgpwd@jira_postgres:5432/jira_db?sslmode=disable"

	var store *pusher.Storage

	for i := 0; i < 15; i++ {
		store, err = pusher.NewStorage(connectionString)
		if err == nil {
			log.Println("Successfully connected to database")
			break
		}
		log.Printf("Database not ready (attempt %d/15), waiting...", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatal(err)
	}

	return &Server{
		configReader: reader,
		config:       serverConfig.NewServerConfig(cfg.Repository, cfg.ConnectorHost, cfg.ConnectorPort),
		storage:      store,
	}
}

func (server *Server) routes() {
	http.HandleFunc("GET /api/v1/projects", server.projects)
	http.HandleFunc("POST /api/v1/updateProject", server.updateProject)
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

	project, err := connector.GetProject(server.config.Repository, projectKey)
	if err != nil {
		http.Error(writer, fmt.Sprintf("error while downloading issues for project %q: %v", projectKey, err), http.StatusBadRequest)

		return
	}

	err = connector.GetIssues(server.config.Repository, project)
	if err != nil {
		http.Error(writer, fmt.Sprintf("error while downloading issues for project %q: %v", projectKey, err), http.StatusBadRequest)
		return
	}

	parsedIssues, err := dataTransformer.ParseIssuesOfProject(project)
	if err != nil {
		http.Error(writer, fmt.Sprintf("error while downloading issues for project %q: %v", projectKey, err), http.StatusBadRequest)
	}

	ctx := context.Background()
	err = server.storage.SaveProject(ctx, parsedIssues, project)
	if err != nil {
		http.Error(writer, fmt.Sprintf("error while saving issues for project %q: %v", projectKey, err), http.StatusBadRequest)
	}
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

func handleProjects(search string) ([]projectModels.Project, error) {
	projects, err := connector.GetProjects("https://issues.apache.org")
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
