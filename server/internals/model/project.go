package model

type Project struct {
	ProjectID   string `json:"project_id"`
	Name        string `json:"name"`
	IssuesCount int    `json:"issues_count"`
}
