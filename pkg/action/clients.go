package action

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v93/github"
	"github.com/promhippie/github_exporter/pkg/config"
	"github.com/promhippie/github_exporter/pkg/exporter"
)

// InstallationClients uses a client per GitHub app installation, resolved by
// the login of the organization or user, or the slug of the enterprise, the
// app has been installed on.
type InstallationClients struct {
	owners   map[string]*github.Client
	fallback *github.Client
}

// Owner returns the client for the given owner, or the default client if
// the app is not installed on it.
func (i *InstallationClients) Owner(owner string) *github.Client {
	if client, ok := i.owners[strings.ToLower(owner)]; ok {
		return client
	}

	return i.fallback
}

// Default returns the client of the first configured owner the app is
// installed on, preferring enterprises over organizations.
func (i *InstallationClients) Default() *github.Client {
	return i.fallback
}

// appInstallation is decoded on its own, as the account of an enterprise
// installation only provides a slug, which is missing on github.User.
type appInstallation struct {
	ID         int64  `json:"id"`
	TargetType string `json:"target_type"`
	Account    struct {
		Login string `json:"login"`
		Slug  string `json:"slug"`
	} `json:"account"`
}

func (a appInstallation) owner() string {
	if a.Account.Slug != "" && strings.EqualFold(a.TargetType, "enterprise") {
		return strings.ToLower(a.Account.Slug)
	}

	return strings.ToLower(a.Account.Login)
}

func useInstallationDiscovery(cfg *config.Config, _ *slog.Logger) bool {
	return cfg.Target.PrivateKey != "" && cfg.Target.AppID != 0 && cfg.Target.InstallID == 0
}

func getClients(cfg *config.Config, logger *slog.Logger) (exporter.Clients, error) {
	if useInstallationDiscovery(cfg, logger) {
		return getInstallationClients(cfg, logger)
	}

	client, err := getClient(cfg, logger)

	if err != nil {
		return nil, err
	}

	return exporter.StaticClients{
		Client: client,
	}, nil
}

func getInstallationClients(cfg *config.Config, logger *slog.Logger) (exporter.Clients, error) {
	privateKey, err := config.Value(cfg.Target.PrivateKey)

	if err != nil {
		logger.Error("Failed to read GitHub key",
			"err", err,
		)

		return nil, err
	}

	appTransport, err := ghinstallation.NewAppsTransport(
		githubTransport(cfg),
		cfg.Target.AppID,
		[]byte(privateKey),
	)

	if err != nil {
		logger.Error("Failed to create GitHub app transport",
			"err", err,
		)

		return nil, err
	}

	if useEnterprise(cfg, logger) {
		appTransport.BaseURL = enterpriseAPIURL(cfg)
	}

	appClient, err := newClient(cfg, logger, appTransport)

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Target.Timeout)
	defer cancel()

	installations, err := listInstallations(ctx, appClient, cfg.Target.PerPage)

	if err != nil {
		logger.Error("Failed to list GitHub app installations",
			"err", err,
		)

		return nil, err
	}

	result := &InstallationClients{
		owners: make(map[string]*github.Client),
	}

	for _, owner := range configuredOwners(cfg) {
		id, ok := installations[owner]

		if !ok {
			logger.Warn("GitHub app is not installed on owner",
				"owner", owner,
			)

			continue
		}

		client, err := newClient(
			cfg,
			logger,
			ghinstallation.NewFromAppsTransport(appTransport, id),
		)

		if err != nil {
			return nil, err
		}

		logger.Debug("Using GitHub app installation",
			"owner", owner,
			"installation", id,
		)

		result.owners[owner] = client

		if result.fallback == nil {
			result.fallback = client
		}
	}

	if result.fallback == nil {
		return nil, fmt.Errorf("github app %d is not installed on any configured enterprise, org or repo owner", cfg.Target.AppID)
	}

	return result, nil
}

// listInstallations maps the owners the app is installed on to their
// installation ID.
func listInstallations(ctx context.Context, client *github.Client, perPage int) (map[string]int64, error) {
	result := make(map[string]int64)
	page := 1

	for {
		req, err := client.NewRequest(
			ctx,
			"GET",
			fmt.Sprintf("app/installations?per_page=%d&page=%d", perPage, page),
			nil,
		)

		if err != nil {
			return nil, err
		}

		var installations []appInstallation
		resp, err := client.Do(req, &installations)

		if err != nil {
			return nil, err
		}

		for _, installation := range installations {
			result[installation.owner()] = installation.ID
		}

		if resp.NextPage == 0 {
			break
		}

		page = resp.NextPage
	}

	return result, nil
}

// configuredOwners returns the lowercased enterprises, orgs and repo owners
// to collect, in that order and without duplicates.
func configuredOwners(cfg *config.Config) []string {
	result := make([]string, 0)
	seen := make(map[string]bool)

	add := func(owner string) {
		owner = strings.ToLower(strings.TrimSpace(owner))

		if owner == "" || seen[owner] {
			return
		}

		seen[owner] = true
		result = append(result, owner)
	}

	for _, name := range cfg.Target.Enterprises {
		add(name)
	}

	for _, name := range cfg.Target.Orgs {
		add(name)
	}

	for _, name := range cfg.Target.Repos {
		owner, _, _ := strings.Cut(name, "/")
		add(owner)
	}

	return result
}
