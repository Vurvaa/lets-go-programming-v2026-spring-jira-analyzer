CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS authors (
    id SERIAL PRIMARY KEY,
    name text NOT NULL
);

CREATE TABLE IF NOT EXISTS issues (
    id SERIAL PRIMARY KEY,
    project_id INT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    author_id INT NOT NULL REFERENCES authors(id),
    assignee_id INT REFERENCES authors(id),
    key TEXT UNIQUE NOT NULL,
    summary TEXT NOT NULL,
    description TEXT,
    type TEXT,
    priority TEXT,
    status TEXT,
    created_time TIMESTAMP DEFAULT NOW(),
    closed_time TIMESTAMP,
    updated_time TIMESTAMP,
    time_spent INT
);

CREATE TABLE IF NOT EXISTS status_changes (
    id SERIAL PRIMARY KEY,
    issue_id INT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    author_id INT NOT NULL REFERENCES authors(id),
    change_time TIMESTAMP DEFAULT NOW(),
    from_status TEXT NOT NULL,
    to_status TEXT NOT NULL
);