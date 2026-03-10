package projectModels

type PageInfo struct {
	CurrentPage   int `json:"currentPage"`
	PageCount     int `json:"pageCount"`
	ProjectsCount int `json:"projectsCount"`
}

type Project struct {
	ProjectId   string `json:"id"`
	ProjectName string `json:"name"`
}

type ProjectResponse struct {
	Projects []Project `json:"Projects"`
	PageInfo PageInfo  `json:"PageInfo"`
}
