package apiServer

import (
	"context"
	"encoding/json"
	"fmt"
	"jira-connector/internal/logger"
	"net/http"

	"jira-connector/internal/apiServer/handlers"
	"jira-connector/internal/apiServer/models"
	"jira-connector/internal/connector"
	"jira-connector/internal/dataTransformer"
	"jira-connector/internal/pusher"
)

type Server struct {
	repository string
	host       string
	port       uint
	storage    *pusher.Storage
}

func NewServer(cfg models.ServerConfig, store *pusher.Storage) *Server {
	return &Server{
		repository: cfg.Repository,
		host:       cfg.ConnectorHost,
		port:       cfg.ConnectorPort,
		storage:    store,
	}
}

func (server *Server) routes() {
	http.HandleFunc("GET /projects", server.projects)
	http.HandleFunc("POST /updateProject", server.updateProject)
}

func (server *Server) updateProject(writer http.ResponseWriter, request *http.Request) {
	projectKey := request.URL.Query().Get("project")
	if projectKey == "" {
		logger.Instance.Warn("Update request received without project key")
		http.Error(writer, "project name was not passed to /updateProject", http.StatusBadRequest)
		return
	}

	logger.Instance.WithField("project_key", projectKey).Info("Received update request")

	project, err := connector.GetProject(server.repository, projectKey)
	if err != nil {
		logger.Instance.WithError(err).WithField("project_key", projectKey).Error("Failed to fetch project metadata")
		http.Error(writer, fmt.Sprintf("error while downloading issues for project %q: %v", projectKey, err), http.StatusBadRequest)
		return
	}

	err = connector.GetIssues(server.repository, project)
	if err != nil {
		logger.Instance.WithError(err).WithField("project_key", projectKey).Error("Failed to download issues from Jira")
		http.Error(writer, fmt.Sprintf("error while downloading issues for project %q: %v", projectKey, err), http.StatusBadRequest)
		return
	}

	parsedIssues, err := dataTransformer.ParseIssuesOfProject(project)
	if err != nil {
		logger.Instance.WithError(err).WithField("project_key", projectKey).Error("Failed to transform project data")
		http.Error(writer, fmt.Sprintf("error while parsing issues for project %q: %v", projectKey, err), http.StatusBadRequest)
		return
	}

	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write([]byte("Project update started")); err != nil {
		logger.Instance.WithError(err).Error("Error writing response")
		return
	}

	go func() {
		ctx := context.Background()
		if err := server.storage.SaveProject(ctx, parsedIssues, project); err != nil {
			logger.Instance.WithError(err).WithField("project_key", projectKey).Error("Background save failed")
		}
	}()
}

func (server *Server) projects(writer http.ResponseWriter, request *http.Request) {
	limit := handlers.ParseLimit(request)
	page := handlers.ParsePage(request)
	search := handlers.ParseSearch(request)

	projects, err := handlers.HandleProjects(server.repository, search)
	if err != nil {
		logger.Instance.WithError(err).Error("Error downloading list of projects")
		http.Error(writer, fmt.Sprintf("error while downloading list of projects: %v", err), http.StatusBadRequest)
		return
	}

	responseProjects := handlers.GetProjectResponse(page, limit, projects)

	writer.Header().Set("Content-Type", "application/json")
	response, err := json.Marshal(responseProjects)
	if err != nil {
		logger.Instance.WithError(err).Error("Error marshalling projects response")
		http.Error(writer, fmt.Sprintf("error while marshalling response: %v", err), http.StatusInternalServerError)
	}

	_, err = writer.Write(response)
	if err != nil {
		logger.Instance.WithError(err).Error("Error writing response to client")
	}
}

func (server *Server) Start() {
	server.routes()

	addr := fmt.Sprintf("%s:%d", server.host, server.port)
	logger.Instance.WithField("port", server.port).Info("Starting Jira Connector server")

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		logger.Instance.WithError(err).Fatal("Failed to start server")
	}
}
