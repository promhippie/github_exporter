package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v79/github"
	"github.com/jmoiron/sqlx"
)

// storeWorkflowJobEvent handles workflow_run events from GitHub.
func storeWorkflowJobEvent(handle *sqlx.DB, event *github.WorkflowJobEvent) error {
	job := event.WorkflowJob

	record := &WorkflowJob{
		Owner:           event.GetRepo().GetOwner().GetLogin(),
		Repo:            event.GetRepo().GetName(),
		Name:            job.GetName(),
		Status:          job.GetStatus(),
		Conclusion:      job.GetConclusion(),
		Branch:          job.GetHeadBranch(),
		SHA:             job.GetHeadSHA(),
		Identifier:      event.GetWorkflowJob().GetID(),
		RunID:           job.GetRunID(),
		RunAttempt:      int(job.GetRunAttempt()),
		CreatedAt:       job.GetCreatedAt().Unix(),
		StartedAt:       job.GetStartedAt().Unix(),
		CompletedAt:     job.GetCompletedAt().Unix(),
		Labels:          strings.Join(job.Labels, ","),
		RunnerID:        job.GetRunnerID(),
		RunnerName:      job.GetRunnerName(),
		RunnerGroupID:   job.GetRunnerGroupID(),
		RunnerGroupName: job.GetRunnerGroupName(),
		WorkflowName:    job.GetWorkflowName(),
	}

	return createOrUpdateWorkflowJob(handle, record)
}

// createOrUpdateWorkflowJob creates or updates the record.
func createOrUpdateWorkflowJob(handle *sqlx.DB, record *WorkflowJob) error {
	existing := &WorkflowJob{}
	stmt, err := handle.PrepareNamed(findWorkflowJobQuery)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to prepare find: %w", err)
	}

	if err := stmt.Get(existing, record); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to find record: %w", err)
	}

	if existing.Identifier == 0 {
		if _, err := handle.NamedExec(
			createWorkflowJobQuery,
			record,
		); err != nil {
			return fmt.Errorf("failed to create record: %w", err)
		}
	} else {
		if existing.CreatedAt > record.CreatedAt {
			return nil
		} else if existing.CreatedAt == record.CreatedAt && existing.Status == "completed" {
			// The updatedAt timestamp is in seconds, so if the existing record has
			// the same timestamp as the new record, and the status is "completed",
			// we can safely ignore the update.
			return nil
		}

		if _, err := handle.NamedExec(
			updateWorkflowJobQuery,
			record,
		); err != nil {
			return fmt.Errorf("failed to update record: %w", err)
		}
	}

	return nil
}

// getWorkflowJobs retrieves the workflow jobs from the database.
func getWorkflowJobs(handle *sqlx.DB, window time.Duration) ([]*WorkflowJob, error) {
	records := make([]*WorkflowJob, 0)

	rows, err := handle.NamedQuery(
		selectWorkflowJobsQuery,
		map[string]interface{}{
			"window": time.Now().Add(-window).Unix(),
		},
	)

	if err != nil {
		return records, err
	}

	defer func() { _ = rows.Close() }()

	for rows.Next() {
		record := &WorkflowJob{}

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

// pruneWorkflowJobs prunes older workflow job records.
func pruneWorkflowJobs(handle *sqlx.DB, timeframe time.Duration) error {
	if _, err := handle.NamedExec(
		purgeWorkflowJobsQuery,
		map[string]interface{}{
			"timeframe": time.Now().Add(-timeframe).Unix(),
		},
	); err != nil {
		return fmt.Errorf("failed to prune workflow jobs: %w", err)
	}

	return nil
}

var selectWorkflowJobsQuery = `
SELECT
	owner,
	repo,
	name,
	status,
	conclusion,
	branch,
	sha,
	identifier,
	run_id,
	run_attempt,
	created_at,
	started_at,
	completed_at,
	labels,
	runner_id,
	runner_name,
	runner_group_id,
	runner_group_name,
	workflow_name
FROM
	workflow_jobs
WHERE
	created_at > :window
ORDER BY
	created_at ASC;`

var findWorkflowJobQuery = `
SELECT
	identifier
FROM
	workflow_jobs
WHERE
	owner=:owner AND repo=:repo AND identifier=:identifier;`

var createWorkflowJobQuery = `
INSERT INTO workflow_jobs (
	owner,
	repo,
	name,
	status,
	conclusion,
	branch,
	sha,
	identifier,
	run_id,
	run_attempt,
	created_at,
	started_at,
	completed_at,
	labels,
	runner_id,
	runner_name,
	runner_group_id,
	runner_group_name,
	workflow_name
) VALUES (
	:owner,
	:repo,
	:name,
	:status,
	:conclusion,
	:branch,
	:sha,
	:identifier,
	:run_id,
	:run_attempt,
	:created_at,
	:started_at,
	:completed_at,
	:labels,
	:runner_id,
	:runner_name,
	:runner_group_id,
	:runner_group_name,
	:workflow_name
);`

var updateWorkflowJobQuery = `
UPDATE
	workflow_jobs
SET
	run_attempt=:run_attempt,
	conclusion=:conclusion,
	name=:name,
	status=:status,
	branch=:branch,
	sha=:sha,
	identifier=:identifier,
	created_at=:created_at,
	started_at=:started_at,
	completed_at=:completed_at,
	runner_id=:runner_id,
	runner_name=:runner_name,
	runner_group_id=:runner_group_id,
	runner_group_name=:runner_group_name
WHERE
	owner=:owner AND repo=:repo AND identifier=:identifier;`

var purgeWorkflowJobsQuery = `
DELETE FROM
	workflow_jobs
WHERE
	created_at < :timeframe;`
