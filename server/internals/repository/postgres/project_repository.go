package postgres

import (
	"database/sql"
	"errors"
	"server/internals/logger"
	"server/internals/model"

	"github.com/sirupsen/logrus"
)

func (repos *DBRepository) GetAllProjects() ([]model.Project, error) {
	projects := make([]model.Project, 0)

	rows, err := repos.db.Query(
		`SELECT
			projects.id,
			projects.project_key,
			projects.name,
			projects.project_url
		FROM projects
		ORDER BY projects.id`,
	)
	if err != nil {
		logger.Instance.WithError(err).Error("Error with querying all projects")
		return projects, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			logger.Instance.Warn("Unable to Close() on rows")
		}
	}(rows)

	for rows.Next() {
		var project model.Project
		err := rows.Scan(&project.ProjectID, &project.ProjectKey, &project.Name, &project.ProjectURL)
		if err != nil {
			logger.Instance.WithError(err).Error("Error on handling query to the database")
			return projects, err
		}
		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		logger.Instance.WithError(err).Error("Rows iteration error")
		return projects, err
	}

	return projects, nil
}

func (repos *DBRepository) ProjectExistsByID(projectID string) (bool, error) {
	var exists bool

	err := repos.db.QueryRow(
		`SELECT EXISTS(
			SELECT 1
			FROM projects
			WHERE id = $1
		)`,
		projectID,
	).Scan(&exists)
	if err != nil {
		logger.Instance.WithFields(logrus.Fields{
			"project_id": projectID,
			"error":      err,
		}).Error("Error while checking project existence")
		return false, err
	}

	return exists, nil
}

func (repos *DBRepository) GetProjectStatsByID(projectID string) (model.ProjectStats, error) {
	var stats model.ProjectStats

	row := repos.db.QueryRow(
		`
		SELECT
			p.id,
			p.project_key,
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
		GROUP BY p.id, p.project_key, p.name
		`,
		projectID,
	)

	err := row.Scan(
		&stats.ProjectID,
		&stats.Key,
		&stats.Name,
		&stats.AllIssuesCount,
		&stats.OpenIssuesCount,
		&stats.CloseIssuesCount,
		&stats.ResolvedIssuesCount,
		&stats.ProgressIssuesCount,
	)
	if err != nil {
		logger.Instance.WithFields(logrus.Fields{
			"project_id": projectID,
			"error":      err,
		}).Error("Error while getting project stats")
		return stats, err
	}

	return stats, nil
}

func (repos *DBRepository) GetReopenedIssuesCount(projectID string) (int, error) {
	var count int

	row := repos.db.QueryRow(
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
		logger.Instance.WithFields(logrus.Fields{
			"project_id": projectID,
			"error":      err,
		}).Error("Error while getting reopened issues count for project")
		return 0, err
	}

	return count, nil
}

func (repos *DBRepository) GetAverageIssuesCount(projectID string) (float64, error) {
	var avg float64

	row := repos.db.QueryRow(
		`
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (closed_time - created_time))), 0)
		FROM issues
		WHERE project_id = $1
		  AND created_time IS NOT NULL
		  AND closed_time IS NOT NULL
		  AND closed_time >= created_time
		`,
		projectID,
	)

	err := row.Scan(&avg)
	if err != nil {
		logger.Instance.WithFields(logrus.Fields{
			"project_id": projectID,
			"error":      err,
		}).Error("Error while getting average issues count for project")
		return 0, err
	}

	return avg, nil
}

func (repos *DBRepository) GetAverageTime(projectID string) (float64, error) {
	var avg float64

	row := repos.db.QueryRow(
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
		logger.Instance.WithFields(logrus.Fields{
			"project_id": projectID,
			"error":      err,
		}).Error("Error while getting average time for project")
		return 0, err
	}

	return avg, nil
}

func (repos *DBRepository) GetProjectIDByKey(projectKey string) (string, error) {
	var projectID string

	err := repos.db.QueryRow(
		`
		SELECT id
		FROM projects
		WHERE project_key = $1
		`,
		projectKey,
	).Scan(&projectID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Instance.WithFields(logrus.Fields{
				"project_key": projectKey,
			}).Info("Project key was not found")
			return "", sql.ErrNoRows
		}

		logger.Instance.WithFields(logrus.Fields{
			"project_key": projectKey,
			"error":       err,
		}).Error("Error while getting project id by key")
		return "", err
	}

	return projectID, nil
}

func (repos *DBRepository) DeleteProjectByID(id string) error {
	query := `DELETE FROM projects WHERE id = $1`
	_, err := repos.db.Exec(query, id)
	if err != nil {
		logger.Instance.WithFields(logrus.Fields{
			"project_id": id,
			"error":      err,
		}).Error("Failed to delete project from DB")
		return err
	}
	return err
}
