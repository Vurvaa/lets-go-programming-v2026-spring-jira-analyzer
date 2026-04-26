package pusher

import (
	"database/sql"
	"fmt"
	"jira-connector/internal/connector"
	"jira-connector/internal/dataTransformer"
	"log"
)

func (s *Storage) SaveProject(issues []dataTransformer.Issue, project *connector.Project) error {
	log.Printf("Saving %d issues", len(issues))

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	if err := s.upsertProject(
		tx, project.ProjectId,
		project.ProjectKey,
		project.ProjectName,
		project.ProjectSelf); err != nil {
		tx.Rollback()
		return fmt.Errorf("error of saving project: %w", err)
	}

	log.Printf("Starting save %d issues", len(issues))
	for i, issue := range issues {
		if i%100 == 0 {
			log.Printf("Processing issues: %d/%d...", i, len(issues))
		}

		authorName := issue.Fields.Creator.AuthorName
		if err := s.upsertAuthor(tx, authorName); err != nil {
			tx.Rollback()
			return fmt.Errorf("error of saving author: %w", err)
		}
		assigneeName := issue.Fields.Assignee.AssigneeName
		if err := s.upsertAuthor(tx, assigneeName); err != nil {
			tx.Rollback()
			return fmt.Errorf("error of saving assignee: %w", err)
		}

		if err := s.upsertIssue(tx, issue); err != nil {
			tx.Rollback()
			return fmt.Errorf("error of saving issue: %w", err)
		}

		changes := issue.StatusChanges
		issueID := issue.IssueID
		if err := s.upsertStatusChanges(tx, issueID, changes); err != nil {
			tx.Rollback()
			return fmt.Errorf("error of saving status changes: %w", err)
		}
	}

	err = tx.Commit()
	if err == nil {
		log.Println("Saving completed successfully")
	}
	return err
}

func (s *Storage) upsertProject(tx *sql.Tx, projectID, projectKey, projectName, URL string) error {
	query := `INSERT INTO projects (id, project_key, name, project_url) VALUES ($1, $2, $3, $4) ON CONFLICT (id)DO UPDATE SET name = EXCLUDED.name`
	_, err := tx.Exec(query, projectID, projectKey, projectName, URL)
	return err
}

func (s *Storage) upsertAuthor(tx *sql.Tx, name string) error {
	query := `INSERT INTO authors (name) VALUES ($1) ON CONFLICT (name) DO NOTHING`
	_, err := tx.Exec(query, name)
	return err
}

func (s *Storage) upsertStatusChanges(tx *sql.Tx, issueID string, changes []dataTransformer.StatusChanges) error {
	query := `
        INSERT INTO status_changes (id, issue_id, author_name, from_status, to_status, change_time) 
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (id) DO UPDATE SET
            from_status = EXCLUDED.from_status,
            to_status = EXCLUDED.to_status`

	for _, statusChange := range changes {
		if err := s.upsertAuthor(tx, statusChange.AuthorName); err != nil {
			return fmt.Errorf("error of pre-saving status change author: %w", err)
		}

		_, err := tx.Exec(query,
			statusChange.Id,
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

func (s *Storage) upsertIssue(tx *sql.Tx, issue dataTransformer.Issue) error {
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
    	summary = EXCLUDED.summary,
    	assignee_name = EXCLUDED.assignee_name,
    	priority = EXCLUDED.priority,
    	description = EXCLUDED.description,
			closed_time = EXCLUDED.closed_time,
			type = EXCLUDED.type`

	closedTime := sql.NullString{
		String: issue.Fields.ClosedTime,
		Valid:  issue.Fields.ClosedTime != "",
	}

	_, err := tx.Exec(query,
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
