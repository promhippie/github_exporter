package store

import "testing"

func TestEncodeCustomProperties(t *testing.T) {
	t.Run("nil map", func(t *testing.T) {
		if got := encodeCustomProperties(nil); got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})

	t.Run("empty map", func(t *testing.T) {
		if got := encodeCustomProperties(map[string]any{}); got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})

	t.Run("populated map round-trips through ByLabel", func(t *testing.T) {
		encoded := encodeCustomProperties(map[string]any{
			"department": "javateam",
		})

		record := &WorkflowJob{CustomProperties: encoded}

		if got := record.ByLabel("department"); got != "javateam" {
			t.Errorf("expected %q, got %q", "javateam", got)
		}
	})
}
