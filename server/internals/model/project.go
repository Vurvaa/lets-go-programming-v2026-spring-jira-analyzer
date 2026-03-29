package model

type Project struct {
	ProjectID   string `json:"project_id"`
	Name        string `json:"name"`
	IssuesCount int    `json:"issues_count"`
}

type ProjectStats struct {
	ProjectID           string
	Key                 string
	Name                string
	AllIssuesCount      int
	OpenIssuesCount     int
	CloseIssuesCount    int
	ResolvedIssuesCount int
	ReopenedIssuesCount int
	ProgressIssuesCount int
	AverageTime         float64
	AverageIssuesCount  float64
}
