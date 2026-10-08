//go:build sqlite

package store

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/go-github/v92/github"
)

func TestWorkflowJobEnvironmentRoundTrip(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	st, err := New("sqlite://file::memory:?cache=shared", logger)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	if _, err := st.Open(); err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer st.Close()

	if err := st.Migrate(); err != nil {
		t.Fatalf("failed to migrate store: %v", err)
	}

	event := &github.WorkflowJobEvent{
		Repo: &github.Repository{
			Owner: &github.User{Login: github.Ptr("acme")},
			Name:  github.Ptr("widgets"),
		},
		WorkflowJob: &github.WorkflowJob{
			ID:         github.Ptr(int64(1)),
			RunID:      github.Ptr(int64(2)),
			Status:     github.Ptr("completed"),
			Conclusion: github.Ptr("success"),
			CreatedAt:  &github.Timestamp{Time: time.Now()},
		},
		Deployment: &github.Deployment{
			Environment: github.Ptr("production"),
		},
	}

	if err := st.StoreWorkflowJobEvent(event); err != nil {
		t.Fatalf("failed to store event: %v", err)
	}

	records, err := st.GetWorkflowJobs(time.Hour)
	if err != nil {
		t.Fatalf("failed to get workflow jobs: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	if got := records[0].ByLabel("environment"); got != "production" {
		t.Errorf("expected %q, got %q", "production", got)
	}

	eventNoEnv := &github.WorkflowJobEvent{
		Repo: &github.Repository{
			Owner: &github.User{Login: github.Ptr("acme")},
			Name:  github.Ptr("widgets"),
		},
		WorkflowJob: &github.WorkflowJob{
			ID:         github.Ptr(int64(3)),
			RunID:      github.Ptr(int64(4)),
			Status:     github.Ptr("completed"),
			Conclusion: github.Ptr("success"),
			CreatedAt:  &github.Timestamp{Time: time.Now()},
		},
	}

	if err := st.StoreWorkflowJobEvent(eventNoEnv); err != nil {
		t.Fatalf("failed to store event without environment: %v", err)
	}

	records, err = st.GetWorkflowJobs(time.Hour)
	if err != nil {
		t.Fatalf("failed to get workflow jobs: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	for _, record := range records {
		if record.Identifier == 3 && record.ByLabel("environment") != "" {
			t.Errorf("expected empty environment for job without deployment, got %q", record.ByLabel("environment"))
		}
	}
}
