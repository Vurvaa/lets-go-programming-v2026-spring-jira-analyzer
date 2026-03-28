package model

type PageInfo struct {
	CurrentPage   int `json:"currentPage"`
	PageCount     int `json:"pageCount"`
	ProjectsCount int `json:"projectsCount"`
}

type ConnectorProject struct {
	ProjectID   string `json:"id"`
	ProjectName string `json:"name"`
	Existence   bool   `json:"existence"`
}

type ProjectResponse struct {
	Projects []ConnectorProject `json:"Projects"`
	PageInfo PageInfo           `json:"PageInfo"`
}
