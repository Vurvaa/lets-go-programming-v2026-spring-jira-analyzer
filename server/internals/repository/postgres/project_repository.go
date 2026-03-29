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
	projects := make([]model.Project, 0)

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

func (r *ProjectRepository) DeleteProjectByID(projectID string) error {
	result, err := r.db.Exec(
		`DELETE FROM projects
		 WHERE id = $1`,
		projectID,
	)
	if err != nil {
		log.Printf("Error while deleting project id %s: %v", projectID, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error while checking deleted rows for project id %s: %v", projectID, err)
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *ProjectRepository) GetProjectStatsByID(projectID string) (model.ProjectStats, error) {
	var stats model.ProjectStats

	row := r.db.QueryRow(
		`
		SELECT
			p.id,
			p.name,

			COUNT(i.id) AS all_issues_count,

			COUNT(CASE
				WHEN LOWER(COALESCE(i.status, '')) = 'open'
				THEN 1
			END) AS open_issues_count,

			COUNT(CASE
				WHEN LOWER(COALESCE(i.status, '')) = 'closed'
				THEN 1
			END) AS close_issues_count,

			COUNT(CASE
				WHEN LOWER(COALESCE(i.status, '')) = 'resolved'
				THEN 1
			END) AS resolved_issues_count,

			COUNT(CASE
				WHEN LOWER(COALESCE(i.status, '')) IN ('in progress', 'progress')
				THEN 1
			END) AS progress_issues_count

		FROM projects p
		LEFT JOIN issues i ON i.project_id = p.id
		WHERE p.id = $1
		GROUP BY p.id, p.name
		`,
		projectID,
	)

	err := row.Scan(
		&stats.ProjectID,
		&stats.Name,
		&stats.AllIssuesCount,
		&stats.OpenIssuesCount,
		&stats.CloseIssuesCount,
		&stats.ResolvedIssuesCount,
		&stats.ProgressIssuesCount,
	)
	if err != nil {
		log.Printf("Error while getting project stats for project id %s: %v", projectID, err)
		return stats, err
	}

	return stats, nil
}

func (r *ProjectRepository) GetReopenedIssuesCount(projectID string) (int, error) {
	var count int

	row := r.db.QueryRow(
		`
		SELECT COUNT(DISTINCT sc.issue_id)
		FROM status_changes sc
		JOIN issues i ON i.id = sc.issue_id
		WHERE i.project_id = $1
		  AND LOWER(sc.to_status) = 'reopened'
		`,
		projectID,
	)

	err := row.Scan(&count)
	if err != nil {
		log.Printf("Error while getting reopened issues count for project id %s: %v", projectID, err)
		return 0, err
	}

	return count, nil
}

func (r *ProjectRepository) GetAverageIssuesCount(projectID string) (float64, error) {
	var avg float64

	row := r.db.QueryRow(
		`
		SELECT COALESCE(COUNT(*)::float / 7.0, 0)
		FROM issues
		WHERE project_id = $1
		  AND created_time >= NOW() - INTERVAL '7 days'
		`,
		projectID,
	)

	err := row.Scan(&avg)
	if err != nil {
		log.Printf("Error while getting average issues count for project id %s: %v", projectID, err)
		return 0, err
	}

	return avg, nil
}

func (r *ProjectRepository) GetAverageTime(projectID string) (float64, error) {
	var avg float64

	row := r.db.QueryRow(
		`
		SELECT COALESCE(AVG(time_spent), 0)
		FROM issues
		WHERE project_id = $1
		  AND time_spent IS NOT NULL
		`,
		projectID,
	)

	err := row.Scan(&avg)
	if err != nil {
		log.Printf("Error while getting average time for project id %s: %v", projectID, err)
		return 0, err
	}

	return avg, nil
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
