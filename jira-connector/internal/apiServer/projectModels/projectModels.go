package projectModels

type PageInfo struct {
	CurrentPage   int `json:"currentPage"`
	PageCount     int `json:"pageCount"`
	ProjectsCount int `json:"projectsCount"`
}

type Project struct {
	ProjectId   string `json:"Id"`
	ProjectName string `json:"Name"`
	Existence   bool   `json:"Existence"`
}

type ProjectResponse struct {
	Projects []Project `json:"data"`
	PageInfo PageInfo  `json:"PageInfo"`
}
