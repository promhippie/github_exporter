package exporter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractStatusFromHTML_WithDataAttribute(t *testing.T) {
	html := `
    <div class="components-section">
      <div class="component-inner-container" data-component-id="8l4ygp009s5s" data-component-status="operational">
        <span class="name">Git Operations</span>
        <span class="component-status">  Operational </span>
      </div>
      <div class="component-inner-container" data-component-id="4230lsnqdsld" data-component-status="partial_outage">
        <span class="name">Webhooks</span>
        <span class="component-status"> Partial Outage </span>
      </div>
    </div>`

	got := extractStatusFromHTML(html)

	if up, ok := got["Git Operations"]; !ok || !up {
		t.Fatalf("expected Git Operations to be up, got ok=%v up=%v", ok, up)
	}
	if up, ok := got["Webhooks"]; !ok || up {
		t.Fatalf("expected Webhooks to be down, got ok=%v up=%v", ok, up)
	}
}

func TestExtractStatusFromHTML_Fallbacks(t *testing.T) {
	html := `
    <div class="components-section">
      <div class="component-inner-container" data-component-id="brv1bkgrwx7q">
        <span class="name">API Requests</span>
        <span class="component-status">Operational</span>
      </div>
      <div class="component-inner-container" data-component-id="kr09ddfgbfsf">
        <span class="name">Issues</span>
        <span class="component-status">Operational</span>
      </div>
      <div class="component-inner-container" data-component-id="hhtssxt0f5v2">
        <span class="name">Pull Requests</span>
        <!-- no status elements present -->
      </div>
    </div>`

	got := extractStatusFromHTML(html)

	if up := got["API Requests"]; !up {
		t.Fatalf("expected API Requests to be up via .component-status fallback")
	}
	if up := got["Issues"]; !up {
		t.Fatalf("expected Issues to be up via component-status text")
	}
	if up := got["Pull Requests"]; up {
		t.Fatalf("expected Pull Requests to be down when status missing")
	}
}

func TestExtractStatusFromHTML_NonOperationalVariants_TextStatuses(t *testing.T) {
	html := `
    <div class="components-section">
      <div class="component-inner-container" data-component-id="a">
        <span class="name">Comp A</span>
        <span class="component-status">Degraded Performance</span>
      </div>
      <div class="component-inner-container" data-component-id="b">
        <span class="name">Comp B</span>
        <span class="component-status">Partial Outage</span>
      </div>
      <div class="component-inner-container" data-component-id="c">
        <span class="name">Comp C</span>
        <span class="component-status">Major Outage</span>
      </div>
      <div class="component-inner-container" data-component-id="d">
        <span class="name">Comp D</span>
        <span class="component-status">Maintenance</span>
      </div>
      <div class="component-inner-container" data-component-id="e">
        <span class="name">Comp E</span>
        <span class="component-status">under maintenance</span>
      </div>
    </div>`

	got := extractStatusFromHTML(html)

	for _, name := range []string{"Comp A", "Comp B", "Comp C", "Comp D", "Comp E"} {
		if up := got[name]; up {
			t.Fatalf("expected %s to be down for non-operational data-component-status", name)
		}
	}
}

func TestExtractStatusFromHTML_UnknownComponentsIgnored(t *testing.T) {
	html := `
    <div class="components-section">
      <div class="component-inner-container">
        <span class="name">Not A Known Component</span>
        <span class="component-status">Operational</span>
      </div>
      <div class="component-inner-container">
        <span class="name">API Requests</span>
        <span class="component-status">Operational</span>
      </div>
    </div>`

	got := extractStatusFromHTML(html)

	if _, ok := got["Not A Known Component"]; ok {
		t.Fatalf("expected unknown component to be ignored (not present)")
	}
	if up, ok := got["API Requests"]; !ok || !up {
		t.Fatalf("expected known component 'API Requests' to be up and present, got ok=%v up=%v", ok, up)
	}
}

func TestExtractStatusFromHTML_TrimsNameAndStatus(t *testing.T) {
	html := `
    <div class="components-section">
      <div class="component-inner-container">
        <span class="name">  Git Operations  </span>
        <span class="component-status">   Operational   </span>
      </div>
    </div>`

	got := extractStatusFromHTML(html)

	if up, ok := got["Git Operations"]; !ok || !up {
		t.Fatalf("expected trimmed name/status to be recognized as up, got ok=%v up=%v", ok, up)
	}
}

func TestExtractStatusFromHTML_NoComponentsSection(t *testing.T) {
	html := `<div><p>No components here</p></div>`
	got := extractStatusFromHTML(html)
	if len(got) != 0 {
		t.Fatalf("expected empty result when components section is missing, got len=%d", len(got))
	}
}

func TestExtractStatusFromHTML_MissingNameSkipped(t *testing.T) {
	html := `
    <div class="components-section">
      <div class="component-inner-container">
        <span class="component-status">Operational</span>
      </div>
      <div class="component-inner-container">
        <span class="name"></span>
        <span class="component-status">Operational</span>
      </div>
      <div class="component-inner-container">
        <span class="name">Pages</span>
        <span class="component-status">Operational</span>
      </div>
    </div>`

	got := extractStatusFromHTML(html)

	if up, ok := got["Pages"]; !ok || !up {
		t.Fatalf("expected 'Pages' to be present and up, got ok=%v up=%v", ok, up)
	}
	// Only the valid known component should be present
	if len(got) != 1 {
		t.Fatalf("expected only one valid component to be captured, got len=%d", len(got))
	}
}

func TestExtractStatusFromHTML_NonOperationalVariants_KnownComponents(t *testing.T) {
	html := `
    <div class="components-section">
      <div class="component-inner-container">
        <span class="name">Issues</span>
        <span class="component-status">Degraded Performance</span>
      </div>
      <div class="component-inner-container">
        <span class="name">Pages</span>
        <span class="component-status">Partial Outage</span>
      </div>
      <div class="component-inner-container">
        <span class="name">Copilot</span>
        <span class="component-status">Maintenance</span>
      </div>
    </div>`

	got := extractStatusFromHTML(html)

	for _, name := range []string{"Issues", "Pages", "Copilot"} {
		up, ok := got[name]
		if !ok {
			t.Fatalf("expected known component %q to be present", name)
		}
		if up {
			t.Fatalf("expected %s to be down for non-operational status text", name)
		}
	}
}
func TestExtractStatusFromFullPageIfPresent(t *testing.T) {
	// Best-effort test: only execute if the provided page source exists.
	// This verifies real-world parsing against the captured HTML.
	// Location relative to this test file: repo root has githubstatus_page_source.html
	root := filepath.Join("..", "..")
	path := filepath.Join(root, "githubstatus_page_source.html")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("skipping full page test, source not found: %v", err)
		return
	}

	got := extractStatusFromHTML(string(data))

	// Validate a subset of expected components are present and up
	expectedUp := []string{
		"Git Operations",
		"Webhooks",
		"API Requests",
		"Issues",
		"Pull Requests",
		"Actions",
		"Packages",
		"Pages",
		"Codespaces",
		"Copilot",
	}

	for _, name := range expectedUp {
		up, ok := got[name]
		if !ok {
			t.Fatalf("expected component %q to be present", name)
		}
		if !up {
			t.Fatalf("expected component %q to be up (operational)", name)
		}
	}

	// Ensure the informational component is ignored
	if _, ok := got["Visit www.githubstatus.com for more information"]; ok {
		t.Fatalf("did not expect informational component in results")
	}
}
