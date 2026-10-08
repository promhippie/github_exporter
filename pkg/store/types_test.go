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

func TestWorkflowJobByLabelCustomProperties(t *testing.T) {
	record := &WorkflowJob{
		Owner:            "promhippie",
		Repo:             "github_exporter",
		CustomProperties: `{"department":"javateam","cost_centers":["cc-1","cc-2"],"internal":true}`,
	}

	t.Run("known label takes precedence", func(t *testing.T) {
		if got := record.ByLabel("owner"); got != "promhippie" {
			t.Errorf("expected %q, got %q", "promhippie", got)
		}
	})

	t.Run("string custom property", func(t *testing.T) {
		if got := record.ByLabel("department"); got != "javateam" {
			t.Errorf("expected %q, got %q", "javateam", got)
		}
	})

	t.Run("multi-select custom property joined with comma", func(t *testing.T) {
		if got := record.ByLabel("cost_centers"); got != "cc-1,cc-2" {
			t.Errorf("expected %q, got %q", "cc-1,cc-2", got)
		}
	})

	t.Run("non-string scalar custom property", func(t *testing.T) {
		if got := record.ByLabel("internal"); got != "true" {
			t.Errorf("expected %q, got %q", "true", got)
		}
	})

	t.Run("missing custom property", func(t *testing.T) {
		if got := record.ByLabel("missing"); got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})

	t.Run("no custom properties stored", func(t *testing.T) {
		empty := &WorkflowJob{}

		if got := empty.ByLabel("department"); got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})

	t.Run("unparsable custom properties", func(t *testing.T) {
		broken := &WorkflowJob{CustomProperties: "not-json"}

		if got := broken.ByLabel("department"); got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})
}

func TestWorkflowRunByLabelCustomProperties(t *testing.T) {
	record := &WorkflowRun{
		Owner:            "promhippie",
		Repo:             "github_exporter",
		CustomProperties: `{"department":"javateam","cost_centers":["cc-1","cc-2"],"internal":true}`,
	}

	t.Run("known label takes precedence", func(t *testing.T) {
		if got := record.ByLabel("owner"); got != "promhippie" {
			t.Errorf("expected %q, got %q", "promhippie", got)
		}
	})

	t.Run("string custom property", func(t *testing.T) {
		if got := record.ByLabel("department"); got != "javateam" {
			t.Errorf("expected %q, got %q", "javateam", got)
		}
	})

	t.Run("multi-select custom property joined with comma", func(t *testing.T) {
		if got := record.ByLabel("cost_centers"); got != "cc-1,cc-2" {
			t.Errorf("expected %q, got %q", "cc-1,cc-2", got)
		}
	})

	t.Run("non-string scalar custom property", func(t *testing.T) {
		if got := record.ByLabel("internal"); got != "true" {
			t.Errorf("expected %q, got %q", "true", got)
		}
	})

	t.Run("missing custom property", func(t *testing.T) {
		if got := record.ByLabel("missing"); got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})

	t.Run("no custom properties stored", func(t *testing.T) {
		empty := &WorkflowRun{}

		if got := empty.ByLabel("department"); got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})

	t.Run("unparsable custom properties", func(t *testing.T) {
		broken := &WorkflowRun{CustomProperties: "not-json"}

		if got := broken.ByLabel("department"); got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})
}
