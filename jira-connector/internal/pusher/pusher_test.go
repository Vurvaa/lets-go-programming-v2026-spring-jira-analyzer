package pusher

import (
	"database/sql"
	"errors"
	"io"
	"os"
	"testing"

	"jira-connector/internal/connector"
	"jira-connector/internal/dataTransformer"
	"jira-connector/internal/logger"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	logger.Instance = logrus.New()
	logger.Instance.SetOutput(io.Discard)
	os.Exit(m.Run())
}

func beginTxWithMock(t *testing.T, db *sql.DB, mock sqlmock.Sqlmock) *sql.Tx {
	mock.ExpectBegin()
	tx, err := db.Begin()
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	return tx
}

func TestUpsertAuthor_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	s := &Storage{db: db}
	tx := beginTxWithMock(t, db, mock)

	mock.ExpectExec("INSERT INTO authors").
		WithArgs("John Doe").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = s.upsertAuthor(tx, "John Doe")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpsertAuthor_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	s := &Storage{db: db}
	tx := beginTxWithMock(t, db, mock)

	mock.ExpectExec("INSERT INTO authors").
		WithArgs("Jane Doe").
		WillReturnError(errors.New("db error"))

	err = s.upsertAuthor(tx, "Jane Doe")
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpsertProject_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	s := &Storage{db: db}
	tx := beginTxWithMock(t, db, mock)

	mock.ExpectExec("INSERT INTO projects").
		WithArgs("PROJ-1", "PROJ", "Project Name", "http://example.com").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = s.upsertProject(tx, "PROJ-1", "PROJ", "Project Name", "http://example.com")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpsertProject_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	s := &Storage{db: db}
	tx := beginTxWithMock(t, db, mock)

	mock.ExpectExec("INSERT INTO projects").
		WithArgs("PROJ-2", "PROJ2", "Name", "http://url").
		WillReturnError(errors.New("constraint violation"))

	err = s.upsertProject(tx, "PROJ-2", "PROJ2", "Name", "http://url")
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpsertIssue_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	s := &Storage{db: db}
	tx := beginTxWithMock(t, db, mock)

	issue := dataTransformer.Issue{
		IssueID: "ISSUE-1",
		Fields: struct {
			Project struct {
				ProjectID string `json:"id"`
			} `json:"project"`
			Creator struct {
				AuthorName string `json:"name"`
			} `json:"creator"`
			Assignee struct {
				AssigneeName string `json:"name"`
				Key          string `json:"key"`
			} `json:"assignee"`
			Summary     string `json:"summary"`
			Description string `json:"description"`
			IssueType   struct {
				Type string `json:"name"`
			} `json:"issuetype"`
			Priority struct {
				Name string `json:"name"`
			} `json:"priority"`
			Status struct {
				Name string `json:"name"`
			} `json:"status"`
			CreatedTime string `json:"created"`
			UpdatedTime string `json:"updated"`
			ClosedTime  string `json:"resolutiondate"`
			TimeSpent   int    `json:"timespent"`
		}{
			Project: struct {
				ProjectID string `json:"id"`
			}{ProjectID: "PROJ-1"},
			Creator: struct {
				AuthorName string `json:"name"`
			}{AuthorName: "creator"},
			Assignee: struct {
				AssigneeName string `json:"name"`
				Key          string `json:"key"`
			}{AssigneeName: "assignee", Key: ""},
			Summary:     "Test summary",
			Description: "Test description",
			IssueType: struct {
				Type string `json:"name"`
			}{Type: "Bug"},
			Priority: struct {
				Name string `json:"name"`
			}{Name: "Medium"},
			Status: struct {
				Name string `json:"name"`
			}{Name: "Open"},
			CreatedTime: "2023-01-01T00:00:00Z",
			UpdatedTime: "2023-01-02T00:00:00Z",
			ClosedTime:  "",
			TimeSpent:   0,
		},
		StatusChanges: nil,
	}

	closedTime := sql.NullString{String: issue.Fields.ClosedTime, Valid: issue.Fields.ClosedTime != ""}

	mock.ExpectExec("INSERT INTO issues").
		WithArgs(
			issue.IssueID,
			issue.Fields.Project.ProjectID,
			issue.Fields.Creator.AuthorName,
			issue.Fields.Assignee.AssigneeName,
			issue.Fields.Summary,
			issue.Fields.Description,
			issue.Fields.IssueType.Type,
			issue.Fields.Priority.Name,
			issue.Fields.Status.Name,
			issue.Fields.CreatedTime,
			issue.Fields.UpdatedTime,
			closedTime,
			issue.Fields.TimeSpent,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = s.upsertIssue(tx, issue)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpsertStatusChanges_AuthorError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	s := &Storage{db: db}
	tx := beginTxWithMock(t, db, mock)

	changes := []dataTransformer.StatusChanges{
		{
			Id:         "ch1",
			AuthorName: "author1",
			FromStatus: "Open",
			ToStatus:   "In Progress",
			ChangeTime: "2023-01-01T12:00:00Z",
		},
	}

	mock.ExpectExec("INSERT INTO authors").WithArgs("author1").WillReturnError(errors.New("author insert error"))

	err = s.upsertStatusChanges(tx, "issue-123", changes)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveProject_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	s := &Storage{db: db}

	project := &connector.Project{
		ProjectId:   "PROJ-1",
		ProjectKey:  "PROJ",
		ProjectName: "Test Project",
		ProjectSelf: "http://example.com",
	}

	issues := []dataTransformer.Issue{
		{
			IssueID: "ISSUE-1",
			Fields: struct {
				Project struct {
					ProjectID string `json:"id"`
				} `json:"project"`
				Creator struct {
					AuthorName string `json:"name"`
				} `json:"creator"`
				Assignee struct {
					AssigneeName string `json:"name"`
					Key          string `json:"key"`
				} `json:"assignee"`
				Summary     string `json:"summary"`
				Description string `json:"description"`
				IssueType   struct {
					Type string `json:"name"`
				} `json:"issuetype"`
				Priority struct {
					Name string `json:"name"`
				} `json:"priority"`
				Status struct {
					Name string `json:"name"`
				} `json:"status"`
				CreatedTime string `json:"created"`
				UpdatedTime string `json:"updated"`
				ClosedTime  string `json:"resolutiondate"`
				TimeSpent   int    `json:"timespent"`
			}{
				Project: struct {
					ProjectID string `json:"id"`
				}{ProjectID: "PROJ-1"},
				Creator: struct {
					AuthorName string `json:"name"`
				}{AuthorName: "creator"},
				Assignee: struct {
					AssigneeName string `json:"name"`
					Key          string `json:"key"`
				}{AssigneeName: "assignee", Key: ""},
				Summary: "summary",
				IssueType: struct {
					Type string `json:"name"`
				}{Type: "Task"},
				Priority: struct {
					Name string `json:"name"`
				}{Name: "Medium"},
				Status: struct {
					Name string `json:"name"`
				}{Name: "Open"},
			},
			StatusChanges: []dataTransformer.StatusChanges{},
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO projects").
		WithArgs(project.ProjectId, project.ProjectKey, project.ProjectName, project.ProjectSelf).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO authors").WithArgs("creator").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO authors").WithArgs("assignee").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO issues").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = s.SaveProject(issues, project)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveProject_RollbackOnProjectError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	s := &Storage{db: db}
	project := &connector.Project{ProjectId: "PROJ-1"}
	issues := []dataTransformer.Issue{{IssueID: "ISSUE-1"}}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO projects").WillReturnError(errors.New("project error"))
	mock.ExpectRollback()

	err = s.SaveProject(issues, project)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveProject_RollbackOnIssueError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	s := &Storage{db: db}
	project := &connector.Project{ProjectId: "PROJ-1"}
	issues := []dataTransformer.Issue{
		{
			IssueID: "ISSUE-1",
			Fields: struct {
				Project struct {
					ProjectID string `json:"id"`
				} `json:"project"`
				Creator struct {
					AuthorName string `json:"name"`
				} `json:"creator"`
				Assignee struct {
					AssigneeName string `json:"name"`
					Key          string `json:"key"`
				} `json:"assignee"`
				Summary     string `json:"summary"`
				Description string `json:"description"`
				IssueType   struct {
					Type string `json:"name"`
				} `json:"issuetype"`
				Priority struct {
					Name string `json:"name"`
				} `json:"priority"`
				Status struct {
					Name string `json:"name"`
				} `json:"status"`
				CreatedTime string `json:"created"`
				UpdatedTime string `json:"updated"`
				ClosedTime  string `json:"resolutiondate"`
				TimeSpent   int    `json:"timespent"`
			}{
				Project: struct {
					ProjectID string `json:"id"`
				}{ProjectID: "PROJ-1"},
				Creator: struct {
					AuthorName string `json:"name"`
				}{AuthorName: "c"},
				Assignee: struct {
					AssigneeName string `json:"name"`
					Key          string `json:"key"`
				}{AssigneeName: "a", Key: ""},
			},
			StatusChanges: []dataTransformer.StatusChanges{},
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO projects").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO authors").WithArgs("c").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO authors").WithArgs("a").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO issues").WillReturnError(errors.New("issue error"))
	mock.ExpectRollback()

	err = s.SaveProject(issues, project)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
