package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/jmoiron/sqlx"
)

// storeWorkflowRunEvent handles workflow_run events from GitHub.
func storeWorkflowRunEvent(handle *sqlx.DB, event *github.WorkflowRunEvent) error {
	createdAt := event.GetWorkflowRun().GetCreatedAt().Time.Unix()
	updatedAt := event.GetWorkflowRun().GetUpdatedAt().Time.Unix()
	startedAt := event.GetWorkflowRun().GetRunStartedAt().Time.Unix()

	record := &WorkflowRun{
		Owner:      event.GetRepo().GetOwner().GetLogin(),
		Repo:       event.GetRepo().GetName(),
		WorkflowID: event.GetWorkflowRun().GetWorkflowID(),
		Number:     event.GetWorkflowRun().GetRunNumber(),
		Attempt:    event.GetWorkflowRun().GetRunAttempt(),
		Event:      event.GetWorkflowRun().GetEvent(),
		Name:       event.GetWorkflowRun().GetName(),
		Title:      event.GetWorkflowRun().GetDisplayTitle(),
		Status:     event.GetWorkflowRun().GetConclusion(),
		Branch:     event.GetWorkflowRun().GetHeadBranch(),
		SHA:        event.GetWorkflowRun().GetHeadSHA(),
		Identifier: event.GetWorkflowRun().GetID(),
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		StartedAt:  startedAt,
	}

	if record.Status == "" {
		record.Status = event.GetWorkflowRun().GetStatus()
	}

	return createOrUpdateWorkflowRun(handle, record)
}

// createOrUpdateWorkflowRun creates or updates the record.
func createOrUpdateWorkflowRun(handle *sqlx.DB, record *WorkflowRun) error {
	existing := &WorkflowRun{}
	stmt, err := handle.PrepareNamed(findWorkflowRunQuery)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to prepare find: %w", err)
	}

	if err := stmt.Get(existing, record); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to find record: %w", err)
	}

	if existing.Identifier == 0 {
		if _, err := handle.NamedExec(
			createWorkflowRunQuery,
			record,
		); err != nil {
			return fmt.Errorf("failed to create record: %w", err)
		}
	} else {
		if existing.UpdatedAt > record.UpdatedAt {
			return nil
		} else if existing.UpdatedAt == record.UpdatedAt && existing.Status == "completed" {
			// The updatedAt timestamp is in seconds, so if the existing record has
			// the same timestamp as the new record, and the status is "completed",
			// we can safely ignore the update.
			return nil
		}

		if _, err := handle.NamedExec(
			updateWorkflowRunQuery,
			record,
		); err != nil {
			return fmt.Errorf("failed to update record: %w", err)
		}
	}

	return nil
}

// getWorkflowRuns retrieves the workflow runs from the database.
func getWorkflowRuns(handle *sqlx.DB) ([]*WorkflowRun, error) {
	records := make([]*WorkflowRun, 0)

	rows, err := handle.Queryx(
		selectWorkflowRunsQuery,
	)

	if err != nil {
		return records, err
	}

	defer rows.Close()

	for rows.Next() {
		record := &WorkflowRun{}

		if err := rows.StructScan(
			record,
		); err != nil {
			return records, err
		}

		records = append(
			records,
			record,
		)
	}

	if err := rows.Err(); err != nil {
		return records, err
	}

	return records, nil
}

// pruneWorkflowRuns prunes older workflow run records.
func pruneWorkflowRuns(handle *sqlx.DB, timeframe time.Duration) error {
	if _, err := handle.NamedExec(
		purgeWorkflowRunsQuery,
		map[string]interface{}{
			"timeframe": time.Now().Add(-timeframe).Unix(),
		},
	); err != nil {
		return fmt.Errorf("failed to prune workflow runs: %w", err)
	}

	return nil
}

var selectWorkflowRunsQuery = `
SELECT
	owner,
	repo,
	workflow_id,
	number,
	attempt,
	event,
	name,
	title,
	status,
	branch,
	sha,
	identifier,
	created_at,
	updated_at,
	started_at
FROM
	workflow_runs
ORDER BY
	updated_at ASC;`

var findWorkflowRunQuery = `
SELECT
	identifier,
	updated_at
FROM
	workflow_runs
WHERE
	owner=:owner AND repo=:repo AND workflow_id=:workflow_id AND number=:number;`

var createWorkflowRunQuery = `
INSERT INTO workflow_runs (
	owner,
	repo,
	workflow_id,
	number,
	attempt,
	event,
	name,
	title,
	status,
	branch,
	sha,
	identifier,
	created_at,
	updated_at,
	started_at
) VALUES (
	:owner,
	:repo,
	:workflow_id,
	:number,
	:attempt,
	:event,
	:name,
	:title,
	:status,
	:branch,
	:sha,
	:identifier,
	:created_at,
	:updated_at,
	:started_at
);`

var updateWorkflowRunQuery = `
UPDATE
	workflow_runs
SET
	attempt=:attempt,
	event=:event,
	name=:name,
	title=:title,
	status=:status,
	branch=:branch,
	sha=:sha,
	identifier=:identifier,
	created_at=:created_at,
	updated_at=:updated_at,
	started_at=:started_at
WHERE
	owner=:owner AND repo=:repo AND workflow_id=:workflow_id AND number=:number;`

var purgeWorkflowRunsQuery = `
DELETE FROM
	workflow_runs
WHERE
	updated_at < :timeframe;`
