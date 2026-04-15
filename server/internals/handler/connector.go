package handler

import (
	"net/http"
	"server/internals/logger"
)

// GET /api/v1/connector/projects
func (handler *ProjectHandler) GetAllProjectsFromRepository(writer http.ResponseWriter, request *http.Request) {
	rawQuery := request.URL.RawQuery

	body, err := handler.service.GetAllProjectsFromRepository(rawQuery)
	if err != nil {
		logger.Instance.WithError(err).Error("Error fetching projects from repository")
		http.Error(writer, "Failed to fetch projects from external source", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	if _, err = writer.Write(body); err != nil {
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
	}
}

// POST /api/v1/connector/updateProject
func (handler *ProjectHandler) UpdateProject(writer http.ResponseWriter, request *http.Request) {
	rawQuery := request.URL.RawQuery

	body, err := handler.service.UpdateProject(rawQuery)
	if err != nil {
		logger.Instance.WithError(err).Error("Error updating project in repository")
		http.Error(writer, "Failed to update external project", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if _, err = writer.Write(body); err != nil {
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
	}
}
