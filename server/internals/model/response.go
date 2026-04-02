package model

type PageInfo struct {
	CurrentPage   int `json:"currentPage"`
	PageCount     int `json:"pageCount"`
	ProjectsCount int `json:"projectsCount"`
}

type ConnectorProject struct {
	ProjectID   string `json:"Id"`
	ProjectName string `json:"Name"`
	Existence   bool   `json:"Existence"`
}

type ProjectResponse struct {
	Projects []ConnectorProject `json:"data"`
	PageInfo PageInfo           `json:"PageInfo"`
}
