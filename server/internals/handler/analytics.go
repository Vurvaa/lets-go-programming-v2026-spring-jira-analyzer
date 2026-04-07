package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"server/internals/logger"
	"server/internals/model"
	"strings"
)

// GET /api/v1/graph/get/{taskNumber}
func (handler *ProjectHandler) GetGraphByKey(writer http.ResponseWriter, request *http.Request) {
	num := request.PathValue("taskNumber")
	key := request.URL.Query().Get("project")

	projectData, err := handler.service.GetGraphByProjectKey(num, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(writer, "graph not found", http.StatusNotFound)
			return
		}
		logger.Instance.WithError(err).Error("Error getting graph by key")
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)

		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write(projectData)
	if err != nil {
		logger.Instance.WithError(err).Error("error while writing response")
	}
}

// GET /api/v1/graph/compare/{taskNumber}
func (handler *ProjectHandler) CompareProjects(writer http.ResponseWriter, request *http.Request) {
	num := request.PathValue("taskNumber")
	key := request.URL.Query().Get("project")

	keys := strings.Split(key, ",")
	if len(keys) < 2 || len(keys) > 3 {
		http.Error(writer, "invalid project key", http.StatusBadRequest)
		return
	}

	projects, err := handler.service.CompareProjects(num, keys)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(writer, "graph not found", http.StatusNotFound)
			return
		}
		logger.Instance.WithError(err).Error("Error comparing projects")
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)

		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write(projects)
	if err != nil {
		logger.Instance.WithError(err).Error("error while writing response")
	}
}

// GET /api/v1/graph/isAnalyzed
func (handler *ProjectHandler) HasAnyAnalytics(writer http.ResponseWriter, request *http.Request) {
	key := request.URL.Query().Get("project")

	contains, err := handler.service.HasAnyAnalyticsByProjectKey(key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(writer, "graph not found", http.StatusNotFound)
			return
		}
		logger.Instance.WithError(err).Error("Error getting analytics by project key")
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)

		return
	}

	var body model.IsAnalyzedResponse = model.IsAnalyzedResponse{IsAnalyzed: contains}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(body)
}

// POST /api/v1/graph/make/{taskNumber}
func (handler *ProjectHandler) MakeGraph(writer http.ResponseWriter, request *http.Request) {
	num := request.PathValue("taskNumber")
	key := request.URL.Query().Get("project")

	err := handler.service.MakeGraphByProjectKey(num, key)
	if err != nil {
		logger.Instance.WithError(err).Error("Error making graph by project key")
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)

		return
	}
}

// DELETE /api/v1/graph/delete
func (handler *ProjectHandler) DeleteGraphByProjectKey(writer http.ResponseWriter, request *http.Request) {
	key := request.URL.Query().Get("project")

	writer.WriteHeader(http.StatusAccepted)
	writer.Write([]byte("Deletion started"))

	go func(projectKey string) {
		if err := handler.service.DeleteAllGraphByProjectKey(projectKey); err != nil {
			logger.Instance.WithError(err).Error("Error deleting graph by project key")
		}
	}(key)
}
