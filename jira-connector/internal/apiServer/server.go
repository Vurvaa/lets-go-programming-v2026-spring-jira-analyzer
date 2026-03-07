package apiServer

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"jira-connector/internal/apiServer/serverConfig"
	"jira-connector/internal/configReader"
	"jira-connector/internal/connector"
)

type Server struct {
	configReader *configReader.ConfigReader
	config       *serverConfig.ServerConfig
}

func NewServer() *Server {
	reader := configReader.NewConfigReader("jira-connector/config.yaml")
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

	projectName := request.URL.Query().Get("project")
	if projectName == "" {
		http.Error(writer, "project name was not passed to /updateProject", http.StatusBadRequest)

		return
	}

	issues, err := connector.GetProjectIssues(projectName)
	if err != nil {
		http.Error(writer, fmt.Sprintf("error while downloading issues for project %q: %v", projectName, err), http.StatusBadRequest)

		return
	}

	transformedIssues := server.dataTransformer.TransformIssues(issues)

	databasePusher.PushIssues(transformedIssues)
}

func (server *Server) projects(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		http.Error(writer, "incorrect http method for /projects", http.StatusBadRequest)

		return
	}

	limit, page, search := parseProjectParameters(request)

	projects, err := connector.GetProjects(limit, page, search)
	if err != nil {
		http.Error(writer, fmt.Sprintf("error while downloading list of projects: %v", err), http.StatusBadRequest)

		return
	}

	writer.Header().Set("Content-Type", "application/json")
	response, _ := json.Marshal(projects)
	_, _ = writer.Write(response)
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
