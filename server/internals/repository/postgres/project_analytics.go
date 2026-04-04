package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log"
)

func (r *ProjectRepository) SaveOpenTaskTime(projectID string, data []byte) error {
	_, err := r.db.Exec(
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

func (r *ProjectRepository) SaveTaskPriorityCount(projectID string, data []byte) error {
	_, err := r.db.Exec(
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

func (r *ProjectRepository) GetOpenTaskTime(projectID string) ([]byte, error) {
	var data []byte

	err := r.db.QueryRow(
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

func (r *ProjectRepository) GetTaskPriorityCount(projectID string) ([]byte, error) {
	var data []byte

	err := r.db.QueryRow(
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

func (r *ProjectRepository) DeleteProjectAnalytics(ctx context.Context, projectID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
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

func (r *ProjectRepository) HasAnyAnalytics(ctx context.Context, projectID string) (bool, error) {
	var exists bool

	err := r.db.QueryRowContext(
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
