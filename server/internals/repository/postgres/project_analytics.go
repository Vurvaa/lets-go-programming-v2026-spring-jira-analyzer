package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"server/internals/model"
)

func (repos *DBRepository) SaveOpenTaskTime(projectID string, data []byte) error {
	_, err := repos.db.Exec(
		`
		INSERT INTO open_task_time (project_id, creation_time, data)
		VALUES ($1, NOW(), $2::jsonb)
		ON CONFLICT (project_id)
		DO UPDATE SET
			creation_time = EXCLUDED.creation_time,
			data = EXCLUDED.data
		`,
		projectID,
		data,
	)
	if err != nil {
		log.Printf("error while saving open task time for project %s: %v", projectID, err)
		return err
	}

	return nil
}

func (repos *DBRepository) SaveTaskPriorityCount(projectID string, data []byte) error {
	_, err := repos.db.Exec(
		`
		INSERT INTO task_priority_count (project_id, creation_time, data)
		VALUES ($1, NOW(), $2::jsonb)
		ON CONFLICT (project_id)
		DO UPDATE SET
			creation_time = EXCLUDED.creation_time,
			data = EXCLUDED.data
		`,
		projectID,
		data,
	)
	if err != nil {
		log.Printf("error while saving task priority count for project %s: %v", projectID, err)
		return err
	}

	return nil
}

func (repos *DBRepository) GetOpenTaskTime(projectID string) ([]byte, error) {
	var data []byte

	err := repos.db.QueryRow(
		`
		SELECT data
		FROM open_task_time
		WHERE project_id = $1
		`,
		projectID,
	).Scan(&data)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		log.Printf("error while getting open task time for project %s: %v", projectID, err)
		return nil, err
	}

	return data, nil
}

func (repos *DBRepository) GetTaskPriorityCount(projectID string) ([]byte, error) {
	var data []byte

	err := repos.db.QueryRow(
		`
		SELECT data
		FROM task_priority_count
		WHERE project_id = $1
		`,
		projectID,
	).Scan(&data)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		log.Printf("error while getting task priority count for project %s: %v", projectID, err)
		return nil, err
	}

	return data, nil
}

func (repos *DBRepository) DeleteProjectAnalytics(ctx context.Context, projectID string) error {
	tx, err := repos.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("error while starting transaction for deleting analytics of project %s: %v", projectID, err)
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.ExecContext(
		ctx,
		`
		DELETE FROM open_task_time
		WHERE project_id = $1
		`,
		projectID,
	)
	if err != nil {
		log.Printf("error while deleting open task time for project %s: %v", projectID, err)
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		`
		DELETE FROM task_priority_count
		WHERE project_id = $1
		`,
		projectID,
	)
	if err != nil {
		log.Printf("error while deleting task priority count for project %s: %v", projectID, err)
		return err
	}

	if err = tx.Commit(); err != nil {
		log.Printf("error while committing analytics delete transaction for project %s: %v", projectID, err)
		return err
	}

	return nil
}

func (repos *DBRepository) HasAnyAnalytics(ctx context.Context, projectID string) (bool, error) {
	var exists bool

	err := repos.db.QueryRowContext(
		ctx,
		`
		SELECT
			EXISTS (
				SELECT 1
				FROM open_task_time
				WHERE project_id = $1
			)
			OR
			EXISTS (
				SELECT 1
				FROM task_priority_count
				WHERE project_id = $1
			)
		`,
		projectID,
	).Scan(&exists)

	if err != nil {
		log.Printf("error while checking analytics existence for project %s: %v", projectID, err)
		return false, err
	}

	return exists, nil
}

func (repos *DBRepository) GetIssueOpenTimesByProjectID(projectID string) ([]model.IssueOpenTimeRow, error) {
	rows, err := repos.db.Query(
		`
		SELECT created_time, closed_time
		FROM issues
		WHERE project_id = $1
		  AND created_time IS NOT NULL
		  AND closed_time IS NOT NULL
		  AND closed_time >= created_time
		`,
		projectID,
	)
	if err != nil {
		log.Printf("error while getting issue open times for project %s: %v", projectID, err)
		return nil, err
	}
	defer rows.Close()

	result := make([]model.IssueOpenTimeRow, 0)

	for rows.Next() {
		var row model.IssueOpenTimeRow

		if err := rows.Scan(&row.CreatedTime, &row.ClosedTime); err != nil {
			log.Printf("error while scanning issue open times for project %s: %v", projectID, err)
			return nil, err
		}

		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		log.Printf("rows error while getting issue open times for project %s: %v", projectID, err)
		return nil, err
	}

	return result, nil
}

func (repos *DBRepository) GetIssuePrioritiesByProjectID(projectID string) ([]model.IssuePriorityRow, error) {
	rows, err := repos.db.Query(
		`
		SELECT priority
		FROM issues
		WHERE project_id = $1
		  AND priority IS NOT NULL
		  AND priority <> ''
		`,
		projectID,
	)
	if err != nil {
		log.Printf("error while getting issue priorities for project %s: %v", projectID, err)
		return nil, err
	}
	defer rows.Close()

	result := make([]model.IssuePriorityRow, 0)

	for rows.Next() {
		var row model.IssuePriorityRow

		if err := rows.Scan(&row.Priority); err != nil {
			log.Printf("error while scanning issue priorities for project %s: %v", projectID, err)
			return nil, err
		}

		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		log.Printf("rows error while getting issue priorities for project %s: %v", projectID, err)
		return nil, err
	}

	return result, nil
}
