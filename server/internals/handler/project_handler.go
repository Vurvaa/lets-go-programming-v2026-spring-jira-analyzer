package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"server/internals/service"
	"strconv"
)

type ProjectHandler struct {
	service *service.ProjectService
}

func NewProjectHandler(s *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{service: s}
}

// GET /api/v1/projects
func (h *ProjectHandler) GetAllProjects(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	projects, err := h.service.GetAllProjects(offset, limit)
	if err != nil {
		log.Printf("DEBUG ERROR: %v", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Internal Server Error"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(projects)
}

// GET /api/v1/projects/{id}
func (h *ProjectHandler) GetProjectByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	project, err := h.service.GetProject(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		log.Printf("Error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

// DELETE /api/v1/projects/{id}
func (h *ProjectHandler) DeleteProjectByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.service.DeleteProject(id); err != nil {
		http.Error(w, "Failed to delete", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
