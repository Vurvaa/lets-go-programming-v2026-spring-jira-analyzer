CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT
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
    type TEXT,
    priority TEXT,
    status TEXT,
    created_time TIMESTAMP,
    updated_time TIMESTAMP,
    time_spent INT
);

CREATE TABLE IF NOT EXISTS status_changes (
    id SERIAL PRIMARY KEY,
    issue_id TEXT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    author_name TEXT NOT NULL REFERENCES authors(name),
    change_time TIMESTAMP,
    from_status TEXT NOT NULL,
    to_status TEXT NOT NULL
);