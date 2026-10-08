package store

import (
	"testing"
	"time"

	"github.com/google/go-github/v92/github"
)

func TestUnixTimestamp(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		if got := unixTimestamp(github.Timestamp{}); got != 0 {
			t.Errorf("expected 0 for an unset timestamp, got %d", got)
		}
	})

	t.Run("set value", func(t *testing.T) {
		now := time.Now().Truncate(time.Second)

		if got := unixTimestamp(github.Timestamp{Time: now}); got != now.Unix() {
			t.Errorf("expected %d, got %d", now.Unix(), got)
		}
	})
}

func TestWorkflowJobByLabelEnvironment(t *testing.T) {
	t.Run("environment set", func(t *testing.T) {
		record := &WorkflowJob{Environment: "production"}

		if got := record.ByLabel("environment"); got != "production" {
			t.Errorf("expected %q, got %q", "production", got)
		}
	})

	t.Run("environment unset", func(t *testing.T) {
		record := &WorkflowJob{}

		if got := record.ByLabel("environment"); got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})
}
