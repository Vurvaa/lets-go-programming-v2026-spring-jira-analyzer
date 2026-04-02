package apiServer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"jira-connector/internal/apiServer/handlers"
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

	limit, page, search := handlers.ParseProjectParameters(request)

	projects, err := handlers.HandleProjects(server.config.Repository, search)
	if err != nil {
		http.Error(writer, fmt.Sprintf("error while downloading list of projects: %v", err), http.StatusBadRequest)

		return
	}

	responseProjects := handlers.GetProjectResponse(page, limit, projects)

	writer.Header().Set("Content-Type", "application/json")
	response, err := json.Marshal(responseProjects)
	if err != nil {
		http.Error(writer, fmt.Sprintf("error while marshalling response: %v", err), http.StatusInternalServerError)
	}

	_, _ = writer.Write(response)
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
