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

func (r *ProjectRepository) GetAllProjects() ([]model.Project, error) {
	var projects []model.Project

	rows, err := r.db.Query(
		`SELECT
			projects.id,
			projects.name
		FROM projects
		ORDER BY projects.id`,
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
		err := rows.Scan(&project.ProjectID, &project.Name)
		if err != nil {
			log.Printf("Error on handling query to the database: %s", err.Error())
			return projects, err
		}
		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Rows iteration error: %v", err)
		return projects, err
	}

	return projects, nil
}

func (r *ProjectRepository) ProjectExistsByID(projectID string) (bool, error) {
	var exists bool

	err := r.db.QueryRow(
		`SELECT EXISTS(
			SELECT 1
			FROM projects
			WHERE id = $1
		)`,
		projectID,
	).Scan(&exists)
	if err != nil {
		log.Printf("Error while checking project existence by id %s: %v", projectID, err)
		return false, err
	}

	return exists, nil
}
