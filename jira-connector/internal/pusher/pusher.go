package pusher

import (
	"context"
	"database/sql"
	"fmt"
	"jira-connector/internal/connector"
	"jira-connector/internal/dataTransformer"
	"jira-connector/internal/logger"

	"github.com/sirupsen/logrus"
)

func (s *Storage) SaveProject(ctx context.Context, issues []dataTransformer.Issue, project *connector.Project) error {
	logger.Instance.WithFields(logrus.Fields{
		"project_key": project.ProjectKey,
		"issue_count": len(issues),
	}).Info("Starting to save project data")

	if err := s.upsertProject(
		ctx, project.ProjectId,
		project.ProjectKey,
		project.ProjectName,
		project.ProjectSelf); err != nil {
		logger.Instance.WithError(err).WithField("project_key", project.ProjectKey).Error("Failed to save project")
		return fmt.Errorf("error of saving project: %w", err)
	}

	logger.Instance.WithField("issues_count", len(issues)).Info("Starting save %d issues")
	for i, issue := range issues {
		if i%100 == 0 {
			logger.Instance.WithFields(logrus.Fields{
				"project_key": project.ProjectKey,
				"processed":   i,
				"total":       len(issues),
			}).Info("Processing issues progress...")
		}

		authorName := issue.Fields.Creator.AuthorName
		if err := s.upsertAuthor(ctx, authorName); err != nil {
			logger.Instance.WithError(err).WithField("author", authorName).Error("Failed to save author")
			return fmt.Errorf("error of saving author: %w", err)
		}
		assigneeName := issue.Fields.Assignee.AssigneeName
		if err := s.upsertAuthor(ctx, assigneeName); err != nil {
			logger.Instance.WithError(err).WithField("assignee", assigneeName).Error("Failed to save assignee")
			return fmt.Errorf("error of saving assignee: %w", err)
		}

		if err := s.upsertIssue(ctx, issue); err != nil {
			logger.Instance.WithError(err).WithField("issue_id", issue.IssueID).Error("Failed to save issue")
			return fmt.Errorf("error of saving issue: %w", err)
		}

		changes := issue.StatusChanges
		issueID := issue.IssueID
		if err := s.upsertStatusChanges(ctx, issueID, changes); err != nil {
			logger.Instance.WithError(err).WithField("issue_id", issueID).Error("Failed to save status changes")
			return fmt.Errorf("error of saving status changes: %w", err)
		}
	}

	logger.Instance.Info("Saving completed successfully")
	return nil
}

func (s *Storage) upsertProject(ctx context.Context, projectID, projectKey, projectName, URL string) error {
	query := `INSERT INTO projects (id, project_key, name, project_url) VALUES ($1, $2, $3, $4) ON CONFLICT (id)DO UPDATE SET name = EXCLUDED.name`
	_, err := s.db.ExecContext(ctx, query, projectID, projectKey, projectName, URL)
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
			                    created_time, updated_time, closed_time,
			                    time_spent)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			ON CONFLICT (id) DO UPDATE SET
					status = EXCLUDED.status,
					updated_time = EXCLUDED.updated_time,
					time_spent = EXCLUDED.time_spent,
					summary = EXCLUDED.summary`

	closedTime := sql.NullString{
		String: issue.Fields.ClosedTime,
		Valid:  issue.Fields.ClosedTime != "",
	}

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
		closedTime,                         // $12
		issue.Fields.TimeSpent,             // $13
	)
	if err != nil {
		return fmt.Errorf("error of upserting issue: %w", err)
	}

	return nil
}
