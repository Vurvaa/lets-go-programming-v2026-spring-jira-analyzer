package models

type PageInfo struct {
	CurrentPage   int `json:"currentPage"`
	PageCount     int `json:"pageCount"`
	ProjectsCount int `json:"projectsCount"`
}

type Project struct {
	ProjectId   string `json:"Id"`
	ProjectKey  string `json:"Key"`
	ProjectName string `json:"Name"`
	ProjectUrl  string `json:"Url"`
	Existence   bool   `json:"Existence"`
}

type ProjectResponse struct {
	Projects []Project `json:"data"`
	PageInfo PageInfo  `json:"PageInfo"`
}
