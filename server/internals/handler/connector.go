package handler

import (
	"log"
	"net/http"
)

// GET /api/v1/connector/projects
func (handler *ProjectHandler) GetAllProjectsFromRepository(writer http.ResponseWriter, request *http.Request) {
	rawQuery := request.URL.RawQuery

	body, err := handler.service.GetAllProjectsFromRepository(rawQuery)
	if err != nil {
		log.Printf("Error fetching projects from repository: %v", err)
		http.Error(writer, "Failed to fetch projects from external source", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(body)
}

// POST /api/v1/connector/updateProject
func (handler *ProjectHandler) UpdateProject(writer http.ResponseWriter, request *http.Request) {
	rawQuery := request.URL.RawQuery

	body, err := handler.service.UpdateProject(rawQuery)
	if err != nil {
		log.Printf("Error updating project in repository: %v", err)
		http.Error(writer, "Failed to update external project", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(body)
}
