package handlers

import (
	"io"
	"jira-connector/internal/apiServer/models"
	"jira-connector/internal/config"
	"jira-connector/internal/connector"
	"jira-connector/internal/logger"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	logger.Instance = logrus.New()
	logger.Instance.SetOutput(io.Discard)
	os.Exit(m.Run())
}

func TestParseLimit(t *testing.T) {
	tests := []struct {
		name  string
		limit string
		want  int
	}{
		{"no limit parameter", "", defaultLimit},
		{"valid limit 50", "50", 50},
		{"limit greater than max", "200", maxLimit},
		{"limit equal to max", "100", maxLimit},
		{"limit zero", "0", defaultLimit},
		{"negative limit", "-5", defaultLimit},
		{"non-numeric", "abc", defaultLimit},
		{"empty string", "   ", defaultLimit},
		{"limit 1", "1", 1},
		{"very large number (overflow)", "9999999999999999999999999999999999999", defaultLimit},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.limit == "" {
				req = httptest.NewRequest("GET", "/", nil)
			} else {
				req = httptest.NewRequest("GET", "/?"+url.Values{"limit": {tt.limit}}.Encode(), nil)
			}
			got := ParseLimit(req)
			if got != tt.want {
				t.Errorf("ParseLimit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePage(t *testing.T) {
	tests := []struct {
		name string
		page string
		want int
	}{
		{"no page parameter", "", defaultPage},
		{"valid page 5", "5", 5},
		{"page zero", "0", defaultPage},
		{"negative page", "-3", defaultPage},
		{"non-numeric", "abc", defaultPage},
		{"empty string", "   ", defaultPage},
		{"page 1", "1", 1},
		{"very large number", "9999999999999999999999", defaultPage},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.page == "" {
				req = httptest.NewRequest("GET", "/", nil)
			} else {
				req = httptest.NewRequest("GET", "/?"+url.Values{"page": {tt.page}}.Encode(), nil)
			}
			got := ParsePage(req)
			if got != tt.want {
				t.Errorf("ParsePage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandleProjects(t *testing.T) {
	testParams := &config.ConnectorConfig{
		MinTimeSleep:      10,
		MaxTimeSleep:      20,
		Goroutines:        2,
		IssueInOneRequest: 50,
	}
	connector.InitParameters(testParams)
	httpmock.Activate(t)
	url := "https://issues.apache.org"
	projectsAPI := []connector.Project{
		{ProjectId: "1", ProjectKey: "AAR", ProjectName: "aardvark", ProjectSelf: "https://.../1"},
		{ProjectId: "2", ProjectKey: "ZZZ", ProjectName: "Zoo Keeper", ProjectSelf: "https://.../2"},
		{ProjectId: "3", ProjectKey: "AAA", ProjectName: "Apache ActiveMQ", ProjectSelf: "https://.../3"},
	}
	httpmock.RegisterResponder("GET", url+"/jira/rest/api/2/project",
		httpmock.NewJsonResponderOrPanic(200, projectsAPI))

	tests := []struct {
		name   string
		search string
		want   []models.Project
		hasErr bool
	}{
		{"empty search", "", []models.Project{
			{ProjectId: "1", ProjectKey: "AAR", ProjectName: "aardvark", ProjectUrl: "https://.../1", Existence: false},
			{ProjectId: "2", ProjectKey: "ZZZ", ProjectName: "Zoo Keeper", ProjectUrl: "https://.../2", Existence: false},
			{ProjectId: "3", ProjectKey: "AAA", ProjectName: "Apache ActiveMQ", ProjectUrl: "https://.../3", Existence: false},
		}, false},
		{"case-insensitive substring", "aard", []models.Project{
			{ProjectId: "1", ProjectKey: "AAR", ProjectName: "aardvark", ProjectUrl: "https://.../1", Existence: false},
		}, false},
		{"match middle", "keeper", []models.Project{
			{ProjectId: "2", ProjectKey: "ZZZ", ProjectName: "Zoo Keeper", ProjectUrl: "https://.../2", Existence: false},
		}, false},
		{"no match", "none", []models.Project{}, false},
		{"match with different case", "ACTIVEMQ", []models.Project{
			{ProjectId: "3", ProjectKey: "AAA", ProjectName: "Apache ActiveMQ", ProjectUrl: "https://.../3", Existence: false},
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HandleProjects(url, tt.search)
			if (err != nil) != tt.hasErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestGetProjectResponse(t *testing.T) {
	projects := []models.Project{
		{ProjectId: "1", ProjectName: "A"},
		{ProjectId: "2", ProjectName: "B"},
		{ProjectId: "3", ProjectName: "C"},
		{ProjectId: "4", ProjectName: "D"},
		{ProjectId: "5", ProjectName: "E"},
	}

	tests := []struct {
		name     string
		page     int
		limit    int
		want     []models.Project
		pageInfo models.PageInfo
	}{
		{
			name:     "first page, limit 2",
			page:     1,
			limit:    2,
			want:     projects[0:2],
			pageInfo: models.PageInfo{CurrentPage: 1, PageCount: 3, ProjectsCount: 5},
		},
		{
			name:     "second page, limit 2",
			page:     2,
			limit:    2,
			want:     projects[2:4],
			pageInfo: models.PageInfo{CurrentPage: 2, PageCount: 3, ProjectsCount: 5},
		},
		{
			name:     "last page, not full",
			page:     3,
			limit:    2,
			want:     projects[4:5],
			pageInfo: models.PageInfo{CurrentPage: 3, PageCount: 3, ProjectsCount: 5},
		},
		{
			name:     "page beyond last",
			page:     5,
			limit:    2,
			want:     []models.Project{},
			pageInfo: models.PageInfo{CurrentPage: 5, PageCount: 3, ProjectsCount: 5},
		},
		{
			name:     "zero projects",
			page:     1,
			limit:    10,
			want:     projects[0:5],
			pageInfo: models.PageInfo{CurrentPage: 1, PageCount: 1, ProjectsCount: 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := GetProjectResponse(tt.page, tt.limit, projects)
			if !reflect.DeepEqual(resp.Projects, tt.want) {
				t.Errorf("Projects = %+v, want %+v", resp.Projects, tt.want)
			}
			if resp.PageInfo != tt.pageInfo {
				t.Errorf("PageInfo = %+v, want %+v", resp.PageInfo, tt.pageInfo)
			}
		})
	}
}
