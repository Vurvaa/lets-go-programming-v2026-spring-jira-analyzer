CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    project_key TEXT NOT NULL,
    name TEXT NOT NULL,
    project_url TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS authors (
    name text PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS issues (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    author_name TEXT NOT NULL REFERENCES authors(name),
    assignee_name TEXT REFERENCES authors(name),
    summary TEXT NOT NULL,
    description TEXT,
    type TEXT NOT NULL,
    priority TEXT,
    status TEXT NOT NULL,
    created_time TIMESTAMP NOT NULL,
    updated_time TIMESTAMP NOT NULL,
    closed_time TIMESTAMP,
    time_spent INT
);

CREATE TABLE IF NOT EXISTS status_changes (
    id TEXT PRIMARY KEY,
    issue_id TEXT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    author_name TEXT NOT NULL REFERENCES authors(name),
    change_time TIMESTAMP NOT NULL,
    from_status TEXT NOT NULL,
    to_status TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS open_task_time (
    project_id TEXT PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,
    creation_time TIMESTAMP NOT NULL,
    data JSONB NOT NULL
);

CREATE TABLE IF NOT EXISTS task_priority_count (
    project_id TEXT PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,
    creation_time TIMESTAMP NOT NULL,
    data JSONB NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_issues_project_id ON issues(project_id);
CREATE INDEX IF NOT EXISTS idx_issues_author_name ON issues(author_name);
CREATE INDEX IF NOT EXISTS idx_issues_assignee_name ON issues(assignee_name);
CREATE INDEX IF NOT EXISTS idx_status_changes_issue_id ON status_changes(issue_id);
CREATE INDEX IF NOT EXISTS idx_status_changes_author_name ON status_changes(author_name);
CREATE INDEX IF NOT EXISTS idx_open_task_time_creation ON open_task_time(creation_time);
CREATE INDEX IF NOT EXISTS idx_task_priority_count_creation ON task_priority_count(creation_time);