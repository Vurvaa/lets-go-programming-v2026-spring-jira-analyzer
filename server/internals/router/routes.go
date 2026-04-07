package router

import (
	"net/http"
	"server/internals/handler"
)

type Router struct {
	mux     *http.ServeMux
	handler *handler.ProjectHandler
}

func NewRouter(projectHandler *handler.ProjectHandler) *Router {
	return &Router{mux: http.NewServeMux(), handler: projectHandler}
}

func (router *Router) Handler() http.Handler {
	router.routes()
	return LoggingMiddleware(router.mux)
}

func (router *Router) routes() {
	router.mux.HandleFunc("GET /api/v1/projects", router.handler.GetAllProjectsFromDB)
	router.mux.HandleFunc("GET /api/v1/projects/{id}", router.handler.GetProjectStatsByID)
	router.mux.HandleFunc("DELETE /api/v1/projects/{id}", router.handler.DeleteProjectByID)

	router.mux.HandleFunc("GET /api/v1/connector/projects", router.handler.GetAllProjectsFromRepository)
	router.mux.HandleFunc("POST /api/v1/connector/updateProject", router.handler.UpdateProject)

	router.mux.HandleFunc("GET /api/v1/graph/get/{taskNumber}", router.handler.GetGraphByKey)
	router.mux.HandleFunc("POST /api/v1/graph/make/{taskNumber}", router.handler.MakeGraph)
	router.mux.HandleFunc("DELETE /api/v1/graph/delete", router.handler.DeleteGraphByProjectKey)
	router.mux.HandleFunc("GET /api/v1/graph/isAnalyzed", router.handler.HasAnyAnalytics)
	router.mux.HandleFunc("GET /api/v1/graph/compare/{taskNumber}", router.handler.CompareProjects)
}
