package store

import (
	"strconv"
)

// WorkflowRun defines the type returned by GitHub.
type WorkflowRun struct {
	Owner      string `db:"owner"`
	Repo       string `db:"repo"`
	WorkflowID int64  `db:"workflow_id"`
	Event      string `db:"event"`
	Name       string `db:"name"`
	Title      string `db:"title"`
	Status     string `db:"status"`
	Branch     string `db:"branch"`
	SHA        string `db:"sha"`
	Number     int    `db:"number"`
	Attempt    int    `db:"attempt"`
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
	}

	return ""
}
