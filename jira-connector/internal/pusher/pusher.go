package pusher

import (
	"context"
	"fmt"
	"jira-connector/internal/dataTransformer"
)

func (s *Storage) SaveAll(ctx context.Context, issues []dataTransformer.Issue) error {
	// the method starts a chain of saving data for all tables
	for _, issue := range issues {
		projectID := issue.Fields.Project.ProjectID
		if err := s.upsertProject(ctx, projectID); err != nil {
			return fmt.Errorf("error of saving project: %w", err)
		}

		authorName := issue.Fields.Creator.AuthorName
		if err := s.upsertAuthor(ctx, authorName); err != nil {
			return fmt.Errorf("error of saving author: %w", err)
		}
		assigneeName := issue.Fields.Assignee.AssigneeName
		if err := s.upsertAuthor(ctx, assigneeName); err != nil {
			return fmt.Errorf("error of saving assignee: %w", err)
		}

		if err := s.upsertIssue(ctx, issue); err != nil {
			return fmt.Errorf("error of saving issue: %w", err)
		}

		changes := issue.StatusChanges
		issueID := issue.IssueID
		if err := s.upsertStatusChanges(ctx, issueID, changes); err != nil {
			return fmt.Errorf("error of saving status changes: %w", err)
		}
	}

	return nil
}

func (s *Storage) upsertProject(ctx context.Context, projectID string) error {
	query := `INSERT INTO projects (id)	VALUES ($1) ON CONFLICT (id) DO NOTHING`
	_, err := s.db.ExecContext(ctx, query, projectID)
	return err
}

func (s *Storage) upsertAuthor(ctx context.Context, name string) error {
	query := `INSERT INTO authors (name) VALUES ($1) ON CONFLICT (name) DO NOTHING`
	_, err := s.db.ExecContext(ctx, query, name)
	return err
}

func (s *Storage) upsertStatusChanges(ctx context.Context, issueID string, changes []dataTransformer.StatusChanges) error {
	query := `
			INSERT INTO status_changes (issue_id, author_name, from_status, to_status, change_time) 
			VALUES ($1, $2, $3, $4, $5)`

	for _, statusChange := range changes {
		// The author of changes may not be the author or assignee of the task
		if err := s.upsertAuthor(ctx, statusChange.AuthorName); err != nil {
			return fmt.Errorf("error of pre-saving status change author: %w", err)
		}

		_, err := s.db.ExecContext(ctx, query,
			issueID,
			statusChange.AuthorName,
			statusChange.FromStatus,
			statusChange.ToStatus,
			statusChange.ChangeTime,
		)
		if err != nil {
			return fmt.Errorf("error of upserting status change: %w", err)
		}
	}

	return nil
}

func (s *Storage) upsertIssue(ctx context.Context, issue dataTransformer.Issue) error {
	query := `
			INSERT INTO issues (id, project_id, author_name,
			                    assignee_name, summary, description,
			                    type, priority, status,
			                    created_time, updated_time, time_spent)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			ON CONFLICT (id) DO UPDATE SET
					status = EXCLUDED.status,
					updated_time = EXCLUDED.updated_time,
					time_spent = EXCLUDED.time_spent,
					summary = EXCLUDED.summary`

	_, err := s.db.ExecContext(ctx, query,
		issue.IssueID,                      // $1
		issue.Fields.Project.ProjectID,     // $2
		issue.Fields.Creator.AuthorName,    // $3
		issue.Fields.Assignee.AssigneeName, // $4
		issue.Fields.Summary,               // $5
		issue.Fields.Description,           // $6
		issue.Fields.IssueType.Type,        // $7
		issue.Fields.Priority.Name,         // $8
		issue.Fields.Status.Name,           // $9
		issue.Fields.CreatedTime,           // $10
		issue.Fields.UpdatedTime,           // $11
		issue.Fields.TimeSpent,             // $12
	)
	if err != nil {
		return fmt.Errorf("error of upserting issue: %w", err)
	}

	return nil
}
