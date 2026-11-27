package exporter

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-github/v93/github"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/promhippie/github_exporter/pkg/config"
)

// fakeRunnerAPI serves org runners and runner groups and records the paths
// that got requested.
type fakeRunnerAPI struct {
	mu        sync.Mutex
	requested []string
	runners   string
}

func (f *fakeRunnerAPI) handler(t *testing.T) http.Handler {
	mux := http.NewServeMux()

	respond := func(body string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			f.mu.Lock()
			f.requested = append(f.requested, r.URL.Path)
			f.mu.Unlock()

			_, _ = io.WriteString(w, body)
		}
	}

	mux.HandleFunc("GET /api/v3/orgs/acme/actions/runners", func(w http.ResponseWriter, r *http.Request) {
		respond(f.runners)(w, r)
	})

	mux.HandleFunc("GET /api/v3/orgs/acme/actions/runner-groups", respond(`{"total_count": 3, "runner_groups": [
		{"id": 1, "name": "Default", "inherited": false},
		{"id": 13, "name": "large-runners", "inherited": false},
		{"id": 3, "name": "Enterprise Default", "inherited": true}
	]}`))

	mux.HandleFunc("GET /api/v3/orgs/acme/actions/runner-groups/1/runners", respond(`{"total_count": 1, "runners": [{"id": 101}]}`))
	mux.HandleFunc("GET /api/v3/orgs/acme/actions/runner-groups/13/runners", respond(`{"total_count": 1, "runners": [{"id": 102}]}`))
	mux.HandleFunc("GET /api/v3/orgs/acme/actions/runner-groups/3/runners", func(w http.ResponseWriter, _ *http.Request) {
		t.Errorf("Expected inherited runner groups not to be listed")
		w.WriteHeader(http.StatusNotFound)
	})

	return mux
}

func (f *fakeRunnerAPI) wasRequested(fragment string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, path := range f.requested {
		if strings.Contains(path, fragment) {
			return true
		}
	}

	return false
}

func collectOrgRunnerGroups(t *testing.T, api *fakeRunnerAPI, labels []string) map[string]string {
	server := httptest.NewServer(api.handler(t))
	t.Cleanup(server.Close)

	client, err := github.NewClient(github.WithEnterpriseURLs(server.URL, server.URL))

	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	collector := NewRunnerCollector(
		slog.New(slog.DiscardHandler),
		StaticClients{Client: client},
		nil,
		prometheus.NewCounterVec(prometheus.CounterOpts{Name: "failures"}, []string{"type"}),
		prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "duration"}, []string{"type"}),
		config.Target{
			Orgs:    []string{"acme"},
			Timeout: 5 * time.Second,
			PerPage: 10,
			Runners: config.Runners{Labels: labels},
		},
	)

	ch := make(chan prometheus.Metric, 100)
	collector.Collect(ch)
	close(ch)

	result := make(map[string]string)

	for metric := range ch {
		if !strings.Contains(metric.Desc().String(), "github_runner_org_online") {
			continue
		}

		out := &dto.Metric{}

		if err := metric.Write(out); err != nil {
			t.Fatalf("Failed to write metric: %v", err)
		}

		values := make(map[string]string)

		for _, pair := range out.GetLabel() {
			values[pair.GetName()] = pair.GetValue()
		}

		result[values["name"]] = values["group"]
	}

	return result
}

func TestRunnerCollectorGroupLabel(t *testing.T) {
	api := &fakeRunnerAPI{runners: `{"total_count": 2, "runners": [
		{"id": 101, "name": "default-runner", "status": "online", "busy": false},
		{"id": 102, "name": "blue-runner", "status": "online", "busy": true}
	]}`}

	groups := collectOrgRunnerGroups(t, api, []string{"owner", "name", "group"})

	expected := map[string]string{
		"default-runner": "Default",
		"blue-runner":    "large-runners",
	}

	for name, group := range expected {
		if groups[name] != group {
			t.Errorf("Expected runner %s in group %q, got %q", name, group, groups[name])
		}
	}
}

func TestRunnerCollectorGroupLabelFromRunnerGroupID(t *testing.T) {
	api := &fakeRunnerAPI{runners: `{"total_count": 1, "runners": [
		{"id": 102, "runner_group_id": 13, "name": "blue-runner", "status": "online", "busy": false}
	]}`}

	groups := collectOrgRunnerGroups(t, api, []string{"name", "group"})

	if groups["blue-runner"] != "large-runners" {
		t.Errorf("Expected runner in group large-runners, got %q", groups["blue-runner"])
	}

	if api.wasRequested("/runner-groups/13/runners") {
		t.Errorf("Expected runner group runners not to be listed when the runner group ID is known")
	}
}

func TestRunnerCollectorWithoutGroupLabel(t *testing.T) {
	api := &fakeRunnerAPI{runners: `{"total_count": 1, "runners": [
		{"id": 102, "name": "blue-runner", "status": "online", "busy": false}
	]}`}

	groups := collectOrgRunnerGroups(t, api, config.RunnerLabels())

	if _, ok := groups["blue-runner"]; !ok {
		t.Fatalf("Expected the runner to be collected")
	}

	if api.wasRequested("/runner-groups") {
		t.Errorf("Expected no runner group requests without the group label")
	}
}
