package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	URL "net/url"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
	"jira-connector/internal/logger"
)

type Project struct {
	ProjectId   string `json:"id"`
	ProjectKey  string `json:"key"`
	ProjectName string `json:"name"`
	ProjectSelf string `json:"self"`
	Issues      []json.RawMessage
}

const (
	apiBase string = "/jira/rest/api/2/"
	factor  int    = 2
)

var (
	cp     *config.ConnectorConfig
	client = http.Client{
		Timeout: 180 * time.Second,
	}
)

func InitParameters(cfg *config.ConnectorConfig) {
	cp = cfg
}

func sleep(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func getBody(ctx context.Context, url string, expansion string) ([]byte, error) {
	delay := cp.MinTimeSleep

	logger.Instance.WithField("url", url+apiBase+expansion).Debug("Sending request to Jira")

	for {
		req, err := http.NewRequestWithContext(ctx, "GET", url+apiBase+expansion, nil)
		if err != nil {
			return nil, err
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == 429 {
			if delay > cp.MaxTimeSleep {
				return nil, fmt.Errorf("response failed with rate limit and \nbody: %s", body)
			}
			jiter := min(cp.MaxTimeSleep-delay, rand.Int64N(delay))

			logger.Instance.WithFields(logrus.Fields{
				"delay_ms": delay + jiter,
				"url":      url + apiBase + expansion,
			}).Warn("Rate limit hit, sleeping before retry")

			if err := sleep(ctx, time.Duration(delay+jiter)*time.Millisecond); err != nil {
				return nil, err
			}
			delay *= int64(factor)
			continue
		}
		if resp.StatusCode > 299 {
			return nil, fmt.Errorf("response failed with status code: %d and\nbody: %s", resp.StatusCode, body)
		}
		return body, nil
	}
}

func GetProject(url, key string) (*Project, error) {
	expansion := fmt.Sprintf("project/%s", key)
	body, err := getBody(context.Background(), url, expansion)
	if err != nil {
		return nil, err
	}

	var project Project
	if typeErr := json.Unmarshal(body, &project); typeErr != nil {
		return nil, fmt.Errorf("type mismatch in API response: %w", typeErr)
	}

	return &project, nil
}

func GetProjects(url string) ([]Project, error) {
	expansion := "project"
	body, err := getBody(context.Background(), url, expansion)
	if err != nil {
		return nil, err
	}
	var projects []Project
	if typeErr := json.Unmarshal(body, &projects); typeErr != nil {
		return nil, fmt.Errorf("type mismatch in API response: %w", typeErr)
	}
	return projects, nil
}

func GetIssues(url string, project *Project) error {
	g, ctx := errgroup.WithContext(context.Background())
	escapedProjectName := `"` + URL.QueryEscape(project.ProjectName) + `"`

	logger.Instance.WithField("project_name", project.ProjectName).Info("Fetching issues metadata from Jira")

	body, err := getBody(ctx, url, fmt.Sprintf(
		"search?jql=project=%s&expand=changelog&startAt=0&maxResults=%d",
		escapedProjectName,
		cp.IssueInOneRequest,
	))
	if err != nil {
		return err
	}
	resp := struct {
		Total  int               `json:"total"`
		Issues []json.RawMessage `json:"issues"`
	}{}
	if typeErr := json.Unmarshal(body, &resp); typeErr != nil {
		return fmt.Errorf("type mismatch in API response: %w", typeErr)
	}
	if resp.Total == 0 {
		return nil
	}
	project.Issues = make([]json.RawMessage, 0, resp.Total)
	project.Issues = append(project.Issues, resp.Issues...)
	pages := (resp.Total + cp.IssueInOneRequest - 1) / cp.IssueInOneRequest
	goroutines := min(pages-1, cp.Goroutines)
	jobs := make(chan int, pages-1)
	var mu sync.Mutex
	start := time.Now()
	for i := 0; i < goroutines; i++ {
		g.Go(func() error {
			for startAt := range jobs {
				body, err := getBody(ctx, url, fmt.Sprintf(
					"search?jql=project=%s&expand=changelog&startAt=%d&maxResults=%d",
					escapedProjectName,
					startAt,
					cp.IssueInOneRequest,
				))
				if err != nil {
					logger.Instance.WithError(err).WithField("startAt", startAt).Error("Worker failed to fetch issues body")
					return err
				}
				resp := struct {
					Issues []json.RawMessage `json:"issues"`
				}{}
				if typeErr := json.Unmarshal(body, &resp); typeErr != nil {
					return fmt.Errorf("type mismatch in API response: %w", typeErr)
				}
				mu.Lock()
				project.Issues = append(project.Issues, resp.Issues...)
				mu.Unlock()
			}
			return nil
		})
	}
	for i := 1; i < pages; i++ {
		jobs <- i * cp.IssueInOneRequest
	}
	close(jobs)
	if err := g.Wait(); err != nil {
		return fmt.Errorf("error fetching issues from project: %s %w", project.ProjectName, err)
	}
	logger.Instance.WithFields(logrus.Fields{
		"project_name": project.ProjectName,
		"duration":     time.Since(start).String(),
		"count":        len(project.Issues),
	}).Info("Project issues download completed")
	return nil
}
