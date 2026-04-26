package dataTransformer

import (
	"encoding/json"
	"jira-connector/internal/connector"
	"testing"
)

func TestParseStatusChanges(t *testing.T) {
	jsonData := []byte(`
{
  "changelog": {
    "histories": [
      {
        "id": "1",
        "author": { "name": "alice" },
        "created": "2024-01-01",
        "items": [
          {
            "field": "status",
            "fromString": "OPEN",
            "toString": "DONE"
          },
          {
            "field": "priority",
            "fromString": "LOW",
            "toString": "HIGH"
          }
        ]
      }
    ]
  }
}
`)
	changes, err := parseStatusChanges(jsonData)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	c := changes[0]
	if c.Id != "1" {
		t.Fatalf("expected id=1, got %s", c.Id)
	}
	if c.AuthorName != "alice" {
		t.Fatalf("expected alice, got %s", c.AuthorName)
	}
	if c.FromStatus != "OPEN" || c.ToStatus != "DONE" {
		t.Fatalf("wrong status mapping")
	}
}

func TestParseIssuesOfProject(t *testing.T) {
	raw := json.RawMessage(`
{
  "id": "100",
  "fields": {
    "project": { "id": "P1" },
    "creator": { "name": "bob" },
    "assignee": { "name": "alice", "key": "A1" },
    "summary": "test issue",
    "description": "desc",
    "issuetype": { "name": "Bug" },
    "priority": { "name": "High" },
    "status": { "name": "Done" },
    "created": "2024-01-01",
    "updated": "2024-01-02",
    "resolutiondate": "2024-01-03",
    "timespent": 123
  },
  "changelog": {
    "histories": [
      {
        "id": "1",
        "author": { "name": "alice" },
        "created": "2024-01-01",
        "items": [
          {
            "field": "status",
            "fromString": "OPEN",
            "toString": "DONE"
          }
        ]
      }
    ]
  }
}
`)
	project := &connector.Project{
		Issues: []json.RawMessage{raw},
	}
	issues, err := ParseIssuesOfProject(project)
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	issue := issues[0]
	if issue.IssueID != "100" {
		t.Fatalf("wrong issue id")
	}
	if len(issue.StatusChanges) != 1 {
		t.Fatalf("expected 1 status change")
	}
	if issue.StatusChanges[0].FromStatus != "OPEN" {
		t.Fatalf("wrong status change parsing")
	}
}

func TestParseIssuesOfProject_BadJSON(t *testing.T) {
	raw := json.RawMessage(`{ invalid json }`)

	project := &connector.Project{
		Issues: []json.RawMessage{raw},
	}

	_, err := ParseIssuesOfProject(project)
	if err == nil {
		t.Fatal("expected error")
	}
}
