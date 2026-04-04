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

func (router *Router) routes() {
	router.mux.HandleFunc("GET /api/v1/projects", router.handler.GetAllProjectsFromDB)
	router.mux.HandleFunc("GET /api/v1/projects/{id}", router.handler.GetProjectStatsByID)
	router.mux.HandleFunc("DELETE /api/v1/projects/{id}", router.handler.DeleteProjectByID)
	router.mux.HandleFunc("GET /api/v1/connector/projects", router.handler.GetAllProjectsFromRepository)
	router.mux.HandleFunc("POST /api/v1/connector/updateProject", router.handler.UpdateProject)
}

func (router *Router) Handler() http.Handler {
	return router.mux
}
