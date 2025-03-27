package store

import (
	"strconv"
)

// WorkflowRun defines the type returned by GitHub.
type WorkflowRun struct {
	Owner string `db:"owner"`
	Repo  string `db:"repo"`

	WorkflowID int64  `db:"workflow_id"`
	Event      string `db:"event"`
	Name       string `db:"name"`
	Title      string `db:"title"`
	Status     string `db:"status"`
	Branch     string `db:"branch"`
	SHA        string `db:"sha"`
	Number     int    `db:"number"`
	Attempt    int    `db:"attempt"`
	Actor      string `db:"actor"`
	Identifier int64  `db:"identifier"`
	CreatedAt  int64  `db:"created_at"`
	UpdatedAt  int64  `db:"updated_at"`
	StartedAt  int64  `db:"started_at"`
}

// ByLabel returns values by the defined list of labels.
func (r *WorkflowRun) ByLabel(label string) string {
	switch label {
	case "owner":
		return r.Owner
	case "repo":
		return r.Repo
	case "workflow":
		return strconv.FormatInt(r.WorkflowID, 10)
	case "event":
		return r.Event
	case "name":
		return r.Name
	case "title":
		return r.Name
	case "status":
		return r.Status
	case "branch":
		return r.Branch
	case "sha":
		return r.SHA
	case "number":
		return strconv.Itoa(r.Number)
	case "attempt":
		return strconv.Itoa(r.Attempt)
	case "run":
		return strconv.FormatInt(r.Identifier, 10)
	case "actor":
		return r.Actor
	}

	return ""
}

// WorkflowJob defines the type returned by GitHub.
type WorkflowJob struct {
	Owner string `db:"owner"`
	Repo  string `db:"repo"`

	Name            string `db:"name"`
	Status          string `db:"status"`
	Conclusion      string `db:"conclusion"`
	Branch          string `db:"branch"`
	SHA             string `db:"sha"`
	Identifier      int64  `db:"identifier"`
	RunID           int64  `db:"run_id"`
	RunAttempt      int    `db:"run_attempt"`
	CreatedAt       int64  `db:"created_at"`
	StartedAt       int64  `db:"started_at"`
	CompletedAt     int64  `db:"completed_at"`
	Labels          string `db:"labels"`
	RunnerID        int64  `db:"runner_id"`
	RunnerName      string `db:"runner_name"`
	RunnerGroupID   int64  `db:"runner_group_id"`
	RunnerGroupName string `db:"runner_group_name"`
	WorkflowName    string `db:"workflow_name"`
}

// ByLabel returns values by the defined list of labels.
func (r *WorkflowJob) ByLabel(label string) string {
	switch label {
	case "owner":
		return r.Owner
	case "repo":
		return r.Repo
	case "name":
		return r.Name
	case "title":
		return r.Name
	case "status":
		return r.Status
	case "branch":
		return r.Branch
	case "sha":
		return r.SHA
	case "identifier":
		return strconv.FormatInt(r.Identifier, 10)
	case "run_id":
		return strconv.FormatInt(r.RunID, 10)
	case "run_attempt":
		return strconv.Itoa(r.RunAttempt)
	case "job":
		return strconv.FormatInt(r.Identifier, 10)
	case "labels":
		return r.Labels
	case "runner_id":
		return strconv.FormatInt(r.RunnerID, 10)
	case "runner_name":
		return r.RunnerName
	case "runner_group_id":
		return strconv.FormatInt(r.RunnerGroupID, 10)
	case "runner_group_name":
		return r.RunnerGroupName
	case "workflow_name":
		return r.WorkflowName
	case "conclusion":
		return r.Conclusion
	}

	return ""
}
