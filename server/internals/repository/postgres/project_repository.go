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
	projects := make([]model.Project, 0)

	rows, err := r.db.Query(
		`SELECT
			projects.id,
			projects.name,
			COUNT(issues.id) AS issues_count
		FROM projects
		LEFT JOIN issues ON projects.id = issues.project_id
		GROUP BY projects.id, projects.name
		ORDER BY projects.id
		LIMIT $1
		OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		log.Printf("Error with querying all projects: %v", err)
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
		err := rows.Scan(&project.ProjectID, &project.Name, &project.IssuesCount)
		if err != nil {
			log.Printf("Error on handling query to the database: %s", err.Error())
			return projects, err
		}
		projects = append(projects, project)
	}

	return projects, nil
}

func (r *ProjectRepository) GetByID(id string) (*model.Project, error) {
	query := `SELECT id, name FROM projects WHERE id = $1`
	var project model.Project
	err := r.db.QueryRow(query, id).Scan(&project.ProjectID, &project.Name)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) Delete(id string) error {
	query := `DELETE FROM projects WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
