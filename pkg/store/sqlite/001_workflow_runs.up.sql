CREATE TABLE workflow_runs (
    owner TEXT NOT NULL,
    repo TEXT NOT NULL,
    workflow_id INTEGER NOT NULL,
    number INTEGER NOT NULL,
    attempt INTEGER,
    event TEXT,
    name TEXT,
    title TEXT,
    status TEXT,
    branch TEXT,
    sha TEXT,
    identifier INTEGER,
    created_at DATETIME,
    updated_at DATETIME,
    started_at DATETIME,
    PRIMARY KEY(owner, repo, workflow_id, number)
);
