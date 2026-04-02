package model

type Project struct {
	ProjectID  string `json:"Id"`
	ProjectKey string `json:"Key"`
	Name       string `json:"Name"`
	ProjectURL string `json:"Url"`
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
