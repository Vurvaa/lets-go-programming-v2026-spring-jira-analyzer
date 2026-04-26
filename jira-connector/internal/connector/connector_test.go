package connector

import (
	"bytes"
	"context"
	"errors"
	"jira-connector/internal/config"
	"net/http"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
)

func TestSleep(t *testing.T) {
	t.Run("constext is open while we are sleeping, returns nil", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		if err := sleep(ctx, time.Millisecond*10); err != nil {
			t.Errorf("expected nil, got %v", err)
		}
		cancel()
	})
	t.Run("context was canceled when we are sleeping, returns error", func(t *testing.T) {
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Millisecond)
		if err := sleep(ctx, time.Millisecond*100); !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("expected context.DeadlineExceeded, got %v", err)
		}
	})
}

func TestGetBody(t *testing.T) {
	cp = &config.ConnectorConfig{10, 30, 30, 200}
	origMinTimeSleep := cp.MinTimeSleep
	origMaxTimeSleep := cp.MaxTimeSleep
	origClient := client
	httpmock.Activate(t)
	t.Cleanup(func() {
		client = origClient
		cp.MinTimeSleep = origMinTimeSleep
		cp.MaxTimeSleep = origMaxTimeSleep
	})
	tests := []struct {
		name       string
		url        string
		expansion  string
		ctx        context.Context
		setupMocks func(t *testing.T)
		wantErr    bool
		wantErrIs  error
		wantBody   []byte
	}{
		{
			name:    "invalid URL",
			url:     "cowsay hello!",
			wantErr: true,
		},
		{
			name: "closed context",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			url:       "https://issues.apache.org",
			expansion: "project",
			setupMocks: func(t *testing.T) {
				httpmock.Deactivate()
				t.Cleanup(func() { httpmock.Activate() })
			},
			wantErr:   true,
			wantErrIs: context.Canceled,
		},
		{
			name:      "client timeout",
			ctx:       context.Background(),
			url:       "https://issues.apache.org",
			expansion: "project",
			setupMocks: func(t *testing.T) {
				client = http.Client{Timeout: 1 * time.Nanosecond}
			},
			wantErr: true,
		},
		{
			name:      "429 with delay > maxSleep",
			ctx:       context.Background(),
			url:       "https://issues.apache.org",
			expansion: "project",
			setupMocks: func(t *testing.T) {
				cp.MinTimeSleep = 2000
				cp.MaxTimeSleep = 1000
				httpmock.RegisterResponder("GET", "https://issues.apache.org/jira/rest/api/2/project",
					httpmock.NewStringResponder(429, "Rate limited"))
			},
			wantErr: true,
		},
		{
			name:      "429 and context canceled during sleep",
			ctx:       context.Background(),
			url:       "https://issues.apache.org",
			expansion: "project",
			setupMocks: func(t *testing.T) {
				httpmock.RegisterResponder("GET", "https://issues.apache.org/jira/rest/api/2/project",
					httpmock.NewStringResponder(429, "Rate limited"))
			},
			wantErr:   true,
			wantErrIs: context.Canceled,
		},
		{
			name:      "success on first try",
			ctx:       context.Background(),
			url:       "https://issues.apache.org",
			expansion: "search?jql=project=Accumulo&expand=changelog&startAt=4745&maxResults=1000",
			setupMocks: func(t *testing.T) {
				httpmock.RegisterResponder("GET", "https://issues.apache.org/jira/rest/api/2/search?jql=project=Accumulo&expand=changelog&startAt=4745&maxResults=1000",
					httpmock.NewStringResponder(200, `{"startAt":4745,"maxResults":1000,"total":4745,"issues":[]}`))
			},
			wantErr:  false,
			wantBody: []byte(`{"startAt":4745,"maxResults":1000,"total":4745,"issues":[]}`),
		},
		{
			name:      "status 500 internal error",
			ctx:       context.Background(),
			url:       "https://issues.apache.org",
			expansion: "project",
			setupMocks: func(t *testing.T) {
				httpmock.RegisterResponder("GET", "https://issues.apache.org/jira/rest/api/2/project",
					httpmock.NewStringResponder(500, "internal error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.Reset()
			client = origClient
			cp.MinTimeSleep = origMinTimeSleep
			cp.MaxTimeSleep = origMaxTimeSleep
			if tt.setupMocks != nil {
				tt.setupMocks(t)
			}
			var cancel context.CancelFunc
			ctx := tt.ctx
			if tt.name == "429 and context canceled during sleep" {
				ctx, cancel = context.WithCancel(context.Background())
				time.AfterFunc(20*time.Millisecond, cancel)
				t.Cleanup(cancel)
			}
			body, err := getBody(ctx, tt.url, tt.expansion)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("error = %v, want error type %v", err, tt.wantErrIs)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !bytes.Equal(body, tt.wantBody) {
				t.Errorf("body = %q, want %q", body, tt.wantBody)
			}
		})
	}
}

func TestGetProject(t *testing.T) {
	cp = &config.ConnectorConfig{10, 30, 30, 200}
	httpmock.Activate(t)
	t.Run("getBody error, returns error", func(t *testing.T) {
		httpmock.RegisterResponder("GET", "https://issues.apache.org/jira/rest/api/2/project/AAR",
			httpmock.NewStringResponder(400, "Bad Requst"))
		if _, err := GetProject("https://issues.apache.org", "AAR"); err == nil {
			t.Error("expected connector error: got nil")
		}
	})
	t.Run("apiType error, returns error", func(t *testing.T) {
		httpmock.RegisterResponder("GET", "https://issues.apache.org/jira/rest/api/2/project/AAR",
			httpmock.NewStringResponder(200, `{"id":"12320120","key":123456789,"name":"aardvark","self":"https://issues.apache.org/jira/rest/api/2/project/12320120"}`))
		if _, err := GetProject("https://issues.apache.org", "AAR"); err == nil {
			t.Error("expected apiType error: got nil")
		}
	})
	t.Run("validation test, returns nil", func(t *testing.T) {
		httpmock.RegisterResponder("GET", "https://issues.apache.org/jira/rest/api/2/project/AAR",
			httpmock.NewStringResponder(200, `{"id":"12320120","key":"AAR","name":"aardvark","self":"https://issues.apache.org/jira/rest/api/2/project/12320120"}`))
		project, err := GetProject("https://issues.apache.org", "AAR")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if project == nil {
			t.Fatal("project is nil")
		}
		if project.ProjectId != "12320120" {
			t.Errorf("ProjectId = %q, want %q", project.ProjectId, "12320120")
		}
		if project.ProjectKey != "AAR" {
			t.Errorf("ProjectKey = %q, want %q", project.ProjectKey, "AAR")
		}
		if project.ProjectName != "aardvark" {
			t.Errorf("ProjectName = %q, want %q", project.ProjectName, "aardvark")
		}
		if project.ProjectSelf != "https://issues.apache.org/jira/rest/api/2/project/12320120" {
			t.Errorf("ProjectSelf = %q, want %q", project.ProjectSelf, "https://issues.apache.org/jira/rest/api/2/project/12320120")
		}
		if project.Issues != nil {
			t.Errorf("Issues should be nil, got %v", project.Issues)
		}
	})
}

func TestGetProjects(t *testing.T) {
	cp = &config.ConnectorConfig{10, 30, 30, 200}
	httpmock.Activate(t)
	t.Run("getBody error, returns error", func(t *testing.T) {
		httpmock.RegisterResponder("GET", "https://issues.apache.org/jira/rest/api/2/project",
			httpmock.NewStringResponder(400, "Bad Requst"))
		if _, err := GetProjects("https://issues.apache.org"); err == nil {
			t.Error("expected connector error: got nil")
		}
	})
	t.Run("apiType error, returns error", func(t *testing.T) {
		httpmock.RegisterResponder("GET", "https://issues.apache.org/jira/rest/api/2/project",
			httpmock.NewStringResponder(200, "OK"))
		_, err := GetProjects("https://issues.apache.org")
		if err == nil {
			t.Error("expected apiType error: got nil")
		}
	})
	t.Run("validation test, returns nil", func(t *testing.T) {
		httpmock.RegisterResponder("GET", "https://issues.apache.org/jira/rest/api/2/project",
			httpmock.NewStringResponder(200, `[{"id":"12320120","key":"AAR","name":"aardvark","self":"https://issues.apache.org/jira/rest/api/2/project/12320120"}]`))
		projects, err := GetProjects("https://issues.apache.org")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, project := range projects {
			if project.ProjectId == "" || project.ProjectKey == "" || project.ProjectName == "" || project.ProjectSelf == "" {
				t.Errorf("empty field: id=%q, key=%q, name=%q, self=%q",
					project.ProjectId, project.ProjectKey, project.ProjectName, project.ProjectSelf)
			}
		}
	})
}

func TestGetIssues(t *testing.T) {
	cp = &config.ConnectorConfig{10, 30, 30, 200}
	httpmock.Activate(t)
	t.Run("total = 0 or unmarshalError", func(t *testing.T) {
		url := `https://issues.apache.org/jira/rest/api/2/search?jql=project="Accumulo"&expand=changelog&startAt=0&maxResults=200`
		httpmock.RegisterResponder("GET", url,
			httpmock.NewStringResponder(200, `{"total":"type error"}`))
		if err := GetIssues("https://issues.apache.org", &Project{ProjectName: "Accumulo"}); err == nil {
			t.Error("expected apiType error")
		}
		httpmock.RegisterResponder("GET", url,
			httpmock.NewStringResponder(200, `{"total":0}`))
		if err := GetIssues("https://issues.apache.org", &Project{ProjectName: "Accumulo"}); err != nil {
			t.Error("expected total = 0 nil")
		}
	})
	t.Run("error group test", func(t *testing.T) {
		firstPage := `https://issues.apache.org/jira/rest/api/2/search?jql=project="Accumulo"&expand=changelog&startAt=0&maxResults=200`
		httpmock.RegisterResponder("GET", firstPage,
			httpmock.NewStringResponder(200, `{"total":675}`))
		url := `=~^https://issues.apache.org/jira/rest/api/2/search.*startAt=[1-9][0-9]*.*`
		httpmock.RegisterResponder("GET", url,
			httpmock.NewStringResponder(400, "getBody error in cycle"))
		if err := GetIssues("https://issues.apache.org", &Project{ProjectName: "Accumulo"}); err == nil {
			t.Error("expected g.Whait() error getBody error in cycle")
		}
		httpmock.RegisterResponder("GET", url,
			httpmock.NewStringResponder(200, `{"issues":}`))
		if err := GetIssues("https://issues.apache.org", &Project{ProjectName: "Accumulo"}); err == nil {
			t.Error("expected g.Whait() error unmarshal in cycle")
		}
	})
	t.Run("valid saving test", func(t *testing.T) {
		firstPage := `https://issues.apache.org/jira/rest/api/2/search?jql=project="Accumulo"&expand=changelog&startAt=0&maxResults=200`
		httpmock.RegisterResponder("GET", firstPage,
			httpmock.NewStringResponder(400, "first page error"))
		if err := GetIssues("https://issues.apache.org", &Project{ProjectName: "Accumulo"}); err == nil {
			t.Error("expected first page error")
		}
		httpmock.RegisterResponder("GET", firstPage,
			httpmock.NewStringResponder(200, `{"total":675}`))
		url := `=~^https://issues.apache.org/jira/rest/api/2/search.*startAt=[1-9][0-9]*.*`
		httpmock.RegisterResponder("GET", url,
			httpmock.NewStringResponder(200, `{"issues":[]}`))
		if err := GetIssues("https://issues.apache.org", &Project{ProjectName: "Accumulo"}); err != nil {
			t.Error("unexpected error")
		}
	})
}
