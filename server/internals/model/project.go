package model

type Project struct {
	ProjectID   int
	Name        string
	IssuesCount int
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
