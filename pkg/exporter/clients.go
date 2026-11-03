package exporter

import (
	"github.com/google/go-github/v93/github"
)

// Clients resolves the GitHub client used to talk to an owner, which can be
// an organization, a user or an enterprise.
type Clients interface {
	// Owner returns the client for the given owner.
	Owner(owner string) *github.Client

	// Default returns the client for requests not bound to an owner.
	Default() *github.Client
}

// StaticClients uses the same client for every owner.
type StaticClients struct {
	Client *github.Client
}

// Owner returns the client for the given owner.
func (s StaticClients) Owner(_ string) *github.Client {
	return s.Client
}

// Default returns the client for requests not bound to an owner.
func (s StaticClients) Default() *github.Client {
	return s.Client
}
