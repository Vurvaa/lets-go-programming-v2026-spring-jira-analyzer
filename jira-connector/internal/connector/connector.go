package connector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	URL "net/url"
	"sync"
	"time"
)

type Project struct {
	ProjectId   string `json:"id"`
	ProjectName string `json:"name"`
	Issues      []json.RawMessage
}

const apiBase string = "/jira/rest/api/2/"
const E = 3

func GetProject(url, key string) (*Project, error) {
	expansion := fmt.Sprintf("project/%s", key)
	body, err := GetBody(url, expansion)
	if err != nil {
		return nil, err
	}
	var project Project
	if typeErr := json.Unmarshal(body, &project); typeErr != nil {
		return nil, fmt.Errorf("Type mismatch in API response: %w", typeErr)
	}
	return &project, nil
}

func GetBody(url string, expansion string) ([]byte, error) {
	minTimeSleep := 4000
	maxTimeSleep := 120000
	for {
		resp, err := http.Get(url + apiBase + expansion)
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == 429 {
			if minTimeSleep > maxTimeSleep {
				return nil, fmt.Errorf("Response failed with rate limit and \nbody: %s\n", body)
			}
			time.Sleep(time.Duration(minTimeSleep) * time.Millisecond)
			minTimeSleep = minTimeSleep * E
			continue
		}
		if resp.StatusCode > 299 {
			return nil, fmt.Errorf("Response failed with status code: %d and\nbody: %s\n", resp.StatusCode, body)
		}
		return body, nil
	}
}

func GetProjects(url string) ([]Project, error) {
	expansion := "project"
	body, err := GetBody(url, expansion)
	if err != nil {
		return nil, err
	}
	var projects []Project
	if typeErr := json.Unmarshal(body, &projects); typeErr != nil {
		return nil, fmt.Errorf("Type mismatch in API response: %w", typeErr)
	}
	return projects, nil
}

func GetIssues(url string, project *Project, issueInOneRequest int, goroutines int) error {
	escapedProjectName := `"` + URL.QueryEscape(project.ProjectName) + `"`
	body, err := GetBody(url, fmt.Sprintf(
		"search?jql=project=%s&expand=changelog&startAt=0&maxResults=%d",
		escapedProjectName,
		issueInOneRequest,
	))
	if err != nil {
		return err
	}
	resp := struct {
		Total  int               `json:"total"`
		Issues []json.RawMessage `json:"issues"`
	}{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("type mismatch in API response: %w", err)
	}
	if resp.Total == 0 {
		return nil
	}
	project.Issues = make([]json.RawMessage, 0, resp.Total)
	project.Issues = append(project.Issues, resp.Issues...)
	pages := (resp.Total + issueInOneRequest - 1) / issueInOneRequest
	goroutines = min(pages-1, goroutines)
	jobs := make(chan int, pages-1)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var requestDoCount int
	start := time.Now()
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for startAt := range jobs {
				body, err := GetBody(url, fmt.Sprintf(
					"search?jql=project=%s&expand=changelog&startAt=%d&maxResults=%d",
					escapedProjectName,
					startAt,
					issueInOneRequest,
				))
				if err != nil {
					fmt.Println(err)
					return
				}
				resp := struct {
					Issues []json.RawMessage `json:"issues"`
				}{}
				if err := json.Unmarshal(body, &resp); err != nil {
					return
				}
				mu.Lock()
				project.Issues = append(project.Issues, resp.Issues...)
				requestDoCount++
				mu.Unlock()
			}
		}()
	}
	for i := 1; i < pages; i++ {
		jobs <- i * issueInOneRequest
	}
	close(jobs)
	wg.Wait()
	if requestDoCount != pages-1 {
		return fmt.Errorf(
			"not all pages loaded: got %d expected %d",
			requestDoCount,
			pages-1,
		)
	}
	fmt.Println(project.ProjectName+" was saved for time: ", time.Since(start))
	return nil
}
