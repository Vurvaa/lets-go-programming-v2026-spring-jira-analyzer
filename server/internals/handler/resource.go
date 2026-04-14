package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"server/internals/logger"
)

// GET /api/v1/projects
func (handler *ProjectHandler) GetAllProjectsFromDB(writer http.ResponseWriter, request *http.Request) {
	projects, err := handler.service.GetAllProjectsFromDB()
	if err != nil {
		logger.Instance.WithError(err).Error("DEBUG ERROR")
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(http.StatusInternalServerError)
		err = json.NewEncoder(writer).Encode(map[string]string{"error": "Internal Server Error"})
		if err != nil {
			log.Printf("error with json encode: %v", err)
		}

		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(writer).Encode(projects)
}

// GET /api/v1/projects/{id}
func (handler *ProjectHandler) GetProjectStatsByID(writer http.ResponseWriter, request *http.Request) {
	id := request.PathValue("id")
	project, err := handler.service.GetProjectStatsByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(writer, "Project not found", http.StatusNotFound)
			return
		}
		logger.Instance.WithError(err).Error("Error getting project stats")
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(project)
	if err != nil {
		log.Printf("error with json encode: %v", err)
	}
}

// DELETE /api/v1/projects/{id}
func (handler *ProjectHandler) DeleteProjectByID(writer http.ResponseWriter, request *http.Request) {
	id := request.PathValue("id")

	writer.WriteHeader(http.StatusAccepted)
	if _, err := writer.Write([]byte("Deletion started")); err != nil {
		log.Printf("error while writing to response: %v", err)
	}

	go func(projectID string) {
		if err := handler.service.DeleteProject(projectID); err != nil {
			logger.Instance.WithError(err).Error("Error deleting project in background")
		}
	}(id)
}
