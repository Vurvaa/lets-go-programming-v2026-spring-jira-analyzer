package postgres

import (
	"database/sql"
	"log"
	"server/internals/model"
)

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) GetAllProjects(offset int, limit int) ([]model.Project, error) {
	var projects []model.Project
	rows, err := r.db.Query(
		`SELECT 
    	projects.id, 
    	projects.title, 
    	COUNT(issue.id) AS issues_count
		FROM 
    	projects
		LEFT JOIN 
    	issue ON projects.id = issue.projectId
		GROUP BY 
    	projects.id, 
    	projects.title
		ORDER BY 
    	projects.id
		LIMIT 
    	$1
		OFFSET 
    	$2`,
		limit, offset,
	)

	if err != nil {
		log.Printf("Error with querying all projects.")
		return projects, err
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("Unable to Close() on rows.")
		}
	}(rows)

	for rows.Next() {
		var project model.Project
		err := rows.Scan(
			&project.ProjectID, &project.Title, &project.IssuesCount,
		)
		if err != nil {
			log.Printf("Error on handling query to the database: %s", err.Error())
			return projects, err
		}

		projects = append(projects, project)
	}

	return projects, nil
}
