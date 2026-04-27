package apiServer

import (
	"encoding/json"
	"io"
	"jira-connector/internal/apiServer/models"
	"jira-connector/internal/config"
	"jira-connector/internal/connector"
	"jira-connector/internal/logger"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	logger.Instance = logrus.New()
	logger.Instance.SetOutput(io.Discard)
	os.Exit(m.Run())
}

func Test_projects(t *testing.T) {
	testParams := &config.ConnectorConfig{
		MinTimeSleep:      10,
		MaxTimeSleep:      20,
		Goroutines:        2,
		IssueInOneRequest: 50,
	}
	connector.InitParameters(testParams)
	httpmock.Activate(t)
	apiProjects := []connector.Project{
		{ProjectId: "1", ProjectKey: "AAR", ProjectName: "aardvark", ProjectSelf: "https://.../1"},
		{ProjectId: "2", ProjectKey: "ZZZ", ProjectName: "Zoo Keeper", ProjectSelf: "https://.../2"},
	}
	httpmock.RegisterResponder("GET", "https://issues.apache.org/jira/rest/api/2/project",
		httpmock.NewJsonResponderOrPanic(200, apiProjects))
	server := NewServer(config.ServerConfig{Repository: "https://issues.apache.org"}, nil)
	tests := []struct {
		name           string
		query          string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "default parameters",
			query:          "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "with limit and page",
			query:          "?limit=1&page=2",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid limit",
			query:          "?limit=abc",
			expectedStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/projects"+tt.query, nil)
			w := httptest.NewRecorder()
			server.projects(w, req)
			if w.Code != tt.expectedStatus {
				t.Errorf("got status %d, want %d", w.Code, tt.expectedStatus)
			}
			var resp models.ProjectResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Errorf("response not valid JSON: %v", err)
			}
		})
	}
	t.Run("error from HandleProjects", func(t *testing.T) {
		httpmock.RegisterResponder("GET", "https://issues.apache.org/jira/rest/api/2/project",
			httpmock.NewStringResponder(500, "internal error"))
		req := httptest.NewRequest("GET", "/projects", nil)
		w := httptest.NewRecorder()
		server.projects(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "error while downloading list of projects") {
			t.Errorf("unexpected body: %s", w.Body.String())
		}
	})
}
