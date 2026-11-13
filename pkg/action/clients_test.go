package action

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/promhippie/github_exporter/pkg/config"
	"github.com/promhippie/github_exporter/pkg/exporter"
)

// fakeGitHub serves the app installations, mints a token per installation
// and records which token reached which organization.
type fakeGitHub struct {
	mu     sync.Mutex
	tokens map[string]string
}

func (f *fakeGitHub) handler(t *testing.T) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v3/app/installations", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			t.Errorf("Expected an app JWT, got %q", r.Header.Get("Authorization"))
		}

		switch r.URL.Query().Get("page") {
		case "1":
			w.Header().Set("Link", fmt.Sprintf(`<http://%s/api/v3/app/installations?per_page=1&page=2>; rel="next"`, r.Host))
			_, _ = io.WriteString(w, `[{"id": 11, "target_type": "Organization", "account": {"login": "Acme"}}]`)
		default:
			_, _ = io.WriteString(w, `[
				{"id": 22, "target_type": "Organization", "account": {"login": "other"}},
				{"id": 33, "target_type": "Enterprise", "account": {"slug": "acme-corp", "name": "Acme Corp"}}
			]`)
		}
	})

	mux.HandleFunc("POST /api/v3/app/installations/{id}/access_tokens", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"token":      "token-" + r.PathValue("id"),
			"expires_at": time.Now().Add(time.Hour).Format(time.RFC3339),
		})
	})

	mux.HandleFunc("GET /api/v3/orgs/{org}", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.tokens[r.PathValue("org")] = r.Header.Get("Authorization")
		f.mu.Unlock()

		_, _ = io.WriteString(w, `{}`)
	})

	return mux
}

func testPrivateKey(t *testing.T) string {
	key, err := rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}))
}

func testConfig(t *testing.T, baseURL string) *config.Config {
	cfg := config.Load()
	cfg.Target.BaseURL = baseURL
	cfg.Target.AppID = 1
	cfg.Target.PrivateKey = testPrivateKey(t)
	cfg.Target.Timeout = 5 * time.Second
	cfg.Target.PerPage = 1

	return cfg
}

func TestGetClientsDiscoversInstallations(t *testing.T) {
	fake := &fakeGitHub{tokens: make(map[string]string)}
	server := httptest.NewServer(fake.handler(t))
	defer server.Close()

	cfg := testConfig(t, server.URL)
	cfg.Target.Enterprises = []string{"acme-corp"}
	cfg.Target.Orgs = []string{"acme", "missing"}
	cfg.Target.Repos = []string{"other/*"}

	clients, err := getClients(cfg, slog.New(slog.DiscardHandler))

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	for _, org := range []string{"ACME", "other"} {
		if _, _, err := clients.Owner(org).Organizations.Get(context.Background(), org); err != nil {
			t.Fatalf("Failed to request org %s: %v", org, err)
		}
	}

	expected := map[string]string{
		"ACME":  "token token-11",
		"other": "token token-22",
	}

	for org, token := range expected {
		if fake.tokens[org] != token {
			t.Errorf("Expected org %s to use %q, got %q", org, token, fake.tokens[org])
		}
	}

	if clients.Owner("missing") != clients.Default() {
		t.Errorf("Expected an owner without installation to use the default client")
	}

	if clients.Default() != clients.Owner("acme-corp") {
		t.Errorf("Expected the enterprise installation to be the default client")
	}
}

func TestGetClientsFailsWithoutInstallation(t *testing.T) {
	fake := &fakeGitHub{tokens: make(map[string]string)}
	server := httptest.NewServer(fake.handler(t))
	defer server.Close()

	cfg := testConfig(t, server.URL)
	cfg.Target.Orgs = []string{"missing"}

	if _, err := getClients(cfg, slog.New(slog.DiscardHandler)); err == nil {
		t.Fatalf("Expected an error when the app is not installed on any owner")
	}
}

func TestGetClientsUsesStaticClient(t *testing.T) {
	cfg := config.Load()
	cfg.Target.Token = "token"

	clients, err := getClients(cfg, slog.New(slog.DiscardHandler))

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if _, ok := clients.(exporter.StaticClients); !ok {
		t.Fatalf("Expected static clients, got %T", clients)
	}
}

func TestConfiguredOwners(t *testing.T) {
	cfg := config.Load()
	cfg.Target.Enterprises = []string{"Corp"}
	cfg.Target.Orgs = []string{"acme", "corp"}
	cfg.Target.Repos = []string{"Acme/app", "other/*"}

	got := strings.Join(configuredOwners(cfg), ",")

	if got != "corp,acme,other" {
		t.Errorf("Expected corp,acme,other, got %s", got)
	}
}
