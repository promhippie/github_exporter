package exporter

import (
	"testing"
)

func TestExtractStatusFromJSON_Basic(t *testing.T) {
	data := []byte(`{
        "components": [
            {"name": "Git Operations", "status": "operational"},
            {"name": "Webhooks", "status": "partial_outage"}
        ]
    }`)

	got := extractStatusFromJSON(data)

	if up, ok := got["Git Operations"]; !ok || !up {
		t.Fatalf("expected Git Operations to be up, got ok=%v up=%v", ok, up)
	}
	if up, ok := got["Webhooks"]; !ok || up {
		t.Fatalf("expected Webhooks to be down, got ok=%v up=%v", ok, up)
	}
}

func TestExtractStatusFromJSON_NonOperationalVariants(t *testing.T) {
	data := []byte(`{
        "components": [
            {"name": "Issues", "status": "degraded_performance"},
            {"name": "Pages", "status": "partial_outage"},
            {"name": "Copilot", "status": "major_outage"},
            {"name": "Actions", "status": "maintenance"}
        ]
    }`)

	got := extractStatusFromJSON(data)

	for _, name := range []string{"Issues", "Pages", "Copilot", "Actions"} {
		up, ok := got[name]
		if !ok {
			t.Fatalf("expected known component %q to be present", name)
		}
		if up {
			t.Fatalf("expected %s to be down for non-operational status", name)
		}
	}
}

func TestExtractStatusFromJSON_UnknownComponentsIgnored(t *testing.T) {
	data := []byte(`{
        "components": [
            {"name": "Not A Known Component", "status": "operational"},
            {"name": "API Requests", "status": "operational"}
        ]
    }`)

	got := extractStatusFromJSON(data)

	if _, ok := got["Not A Known Component"]; ok {
		t.Fatalf("expected unknown component to be ignored (not present)")
	}
	if up, ok := got["API Requests"]; !ok || !up {
		t.Fatalf("expected known component 'API Requests' to be up and present, got ok=%v up=%v", ok, up)
	}
}

func TestExtractStatusFromJSON_TrimsNameAndStatus(t *testing.T) {
	data := []byte(`{
        "components": [
            {"name": "  Git Operations  ", "status": "   operational   "}
        ]
    }`)

	got := extractStatusFromJSON(data)
	if up, ok := got["Git Operations"]; !ok || !up {
		t.Fatalf("expected trimmed name/status to be recognized as up, got ok=%v up=%v", ok, up)
	}
}

func TestExtractStatusFromJSON_Empty(t *testing.T) {
	data := []byte(`{"components": []}`)
	got := extractStatusFromJSON(data)
	if len(got) != 0 {
		t.Fatalf("expected empty result when components list is empty, got len=%d", len(got))
	}
}
