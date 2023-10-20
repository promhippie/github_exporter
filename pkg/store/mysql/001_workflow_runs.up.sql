CREATE TABLE workflow_runs (
    owner VARCHAR(255) NOT NULL,
    repo VARCHAR(255) NOT NULL,
    workflow_id INTEGER NOT NULL,
    number INTEGER NOT NULL,
    attempt INTEGER,
    event VARCHAR(255),
    name VARCHAR(255),
    title VARCHAR(255),
    status VARCHAR(255),
    branch VARCHAR(255),
    sha VARCHAR(255),
    identifier INTEGER,
    created_at DATETIME,
    updated_at DATETIME,
    started_at DATETIME,
    PRIMARY KEY(owner, repo, workflow_id, number)
);
