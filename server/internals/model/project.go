package model

type Project struct {
	ProjectID  string `json:"Id"`
	ProjectKey string `json:"Key"`
	Name       string `json:"Name"`
	ProjectURL string `json:"Url"`
}

type ProjectStats struct {
	ProjectID           string  `json:"projectID"`
	Key                 string  `json:"key"`
	Name                string  `json:"name"`
	AllIssuesCount      int     `json:"allIssuesCount"`
	OpenIssuesCount     int     `json:"openIssuesCount"`
	CloseIssuesCount    int     `json:"closeIssuesCount"`
	ResolvedIssuesCount int     `json:"resolvedIssuesCount"`
	ReopenedIssuesCount int     `json:"reopenedIssuesCount"`
	ProgressIssuesCount int     `json:"progressIssuesCount"`
	AverageTime         float64 `json:"averageTime"`
	AverageIssuesCount  float64 `json:"averageIssuesCount"`
}
