package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	URL "net/url"
	"os"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

type Project struct {
	ProjectId   string `json:"id"`
	ProjectName string `json:"name"`
	Issues      []json.RawMessage
}

const (
	apiBase string = "/jira/rest/api/2/"
	factor  int    = 2
)

type Parameters struct {
	MinTimeSleep      int64 `yaml:"minTimeSleep"`
	MaxTimeSleep      int64 `yaml:"maxTimeSleep"`
	Goroutines        int   `yaml:"threadCount"`
	IssueInOneRequest int   `yaml:"issueInOneRequest"`
}

var (
	cp     *Parameters
	client = http.Client{
		Timeout: time.Minute,
	}
	once sync.Once
)

func InitParameters(path string) {
	once.Do(func() {
		data, err := os.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}
		cp = &Parameters{}
		if err = yaml.Unmarshal(data, cp); err != nil {
			log.Fatal(err)
		}
	})
}

func sleep(ctx context.Context, duration time.Duration) error {
	select {
	case <-time.After(duration):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func getBody(ctx context.Context, url string, expansion string) ([]byte, error) {
	delay := cp.MinTimeSleep
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
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == 429 {
			if delay > cp.MaxTimeSleep {
				return nil, fmt.Errorf("Response failed with rate limit and \nbody: %s\n", body)
			}
			jiter := min(cp.MaxTimeSleep-delay, rand.Int64N(delay))
			fmt.Println("sleep for a ", delay+jiter)
			if err := sleep(ctx, time.Duration(delay+jiter)*time.Millisecond); err != nil {
				return nil, err
			}
			delay *= int64(factor)
			continue
		}
		if resp.StatusCode > 299 {
			return nil, fmt.Errorf("Response failed with status code: %d and\nbody: %s\n", resp.StatusCode, body)
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
		return nil, fmt.Errorf("Type mismatch in API response: %w", typeErr)
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
		return nil, fmt.Errorf("Type mismatch in API response: %w", typeErr)
	}
	return projects, nil
}

func GetIssues(url string, project *Project) error {
	g, ctx := errgroup.WithContext(context.Background())
	escapedProjectName := `"` + URL.QueryEscape(project.ProjectName) + `"`
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
		return fmt.Errorf("Type mismatch in API response: %w", typeErr)
	}
	if resp.Total == 0 {
		return nil
	}
	project.Issues = make([]json.RawMessage, 0, resp.Total)
	project.Issues = append(project.Issues, resp.Issues...)
	pages := (resp.Total + cp.IssueInOneRequest - 1) / cp.IssueInOneRequest
	cp.Goroutines = min(pages-1, cp.Goroutines)
	jobs := make(chan int, pages-1)
	var mu sync.Mutex
	start := time.Now()
	for i := 0; i < cp.Goroutines; i++ {
		g.Go(func() error {
			for startAt := range jobs {
				body, err := getBody(ctx, url, fmt.Sprintf(
					"search?jql=project=%s&expand=changelog&startAt=%d&maxResults=%d",
					escapedProjectName,
					startAt,
					cp.IssueInOneRequest,
				))
				if err != nil {
					fmt.Println("Body returned an error")
					return err
				}
				resp := struct {
					Issues []json.RawMessage `json:"issues"`
				}{}
				if typeErr := json.Unmarshal(body, &resp); typeErr != nil {
					return fmt.Errorf("Type mismatch in API response: %w", typeErr)
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
		return fmt.Errorf("Error fetching issues from project: %s %w", project.ProjectName, err)
	}
	fmt.Println(project.ProjectName+" was saved for time: ", time.Since(start))
	return nil
}
