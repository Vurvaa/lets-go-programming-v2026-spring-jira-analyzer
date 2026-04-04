package model

type PageInfo struct {
	CurrentPage   int `json:"currentPage"`
	PageCount     int `json:"pageCount"`
	ProjectsCount int `json:"projectsCount"`
}

type ConnectorProject struct {
	ProjectID   string `json:"Id"`
	ProjectKey  string `json:"Key"`
	ProjectName string `json:"Name"`
	ProjectURL  string `json:"Url"`
	Existence   bool   `json:"Existence"`
}

type ProjectResponse struct {
	Projects []ConnectorProject `json:"data"`
	PageInfo PageInfo           `json:"PageInfo"`
}
