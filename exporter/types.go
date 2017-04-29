package exporter

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/peterhellberg/link"
)

const (
	// timeFormat is required to parse the API timestamps properly.
	timeFormat = "2006-01-02T15:04:05Z"
)

// Owner represents the API owner records.
type Owner struct {
	Login string `json:"login"`
}

// Repo represents the API repo records.
type Repo struct {
	Owner     Owner      `json:"owner"`
	Name      string     `json:"name"`
	Size      float64    `json:"size"`
	Forks     float64    `json:"forks_count"`
	Issues    float64    `json:"open_issues_count"`
	Stars     float64    `json:"stargazers_count"`
	Watchers  float64    `json:"watchers_count"`
	PushedAt  CustomTime `json:"pushed_at"`
	UpdatedAt CustomTime `json:"updated_at"`
}

// Key generates a usable map key for the repo.
func (r *Repo) Key() string {
	return path.Join(r.Owner.Login, r.Name)
}

// Fetch gathers the repo content from the API.
func (r *Repo) Fetch(name string) error {
	res, err := simpleClient().Get(
		fmt.Sprintf("https://api.github.com/repos/%s", name),
	)

	if err != nil {
		return fmt.Errorf("failed to request %s repository. %s", name, err)
	}

	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(r); err != nil {
		return fmt.Errorf("failed to parse %s repository. %s", name, err)
	}

	return nil
}

// Collection represents the API response for lists.
type Collection struct {
	Repos []*Repo
}

// Fetch gathers the collection content from API.
func (c *Collection) Fetch(name string) error {
	url := fmt.Sprintf("https://api.github.com/orgs/%s/repos", name)

	if err := c.pagination(name, url); err != nil {
		return err
	}

	return nil
}

// pagination fetches the records per page from the API.
func (c *Collection) pagination(name, url string) error {
	var (
		newCollection = &Collection{}
	)

	res, err := simpleClient().Get(
		url,
	)

	if err != nil {
		return fmt.Errorf("failed to request %s organization. %s", name, err)
	}

	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(&newCollection.Repos); err != nil {
		return fmt.Errorf("failed to parse %s organization. %s", name, err)
	}

	c.Repos = append(
		c.Repos,
		newCollection.Repos...,
	)

	for _, l := range link.ParseResponse(res) {
		if l.Rel == "next" {
			if err := c.pagination(name, l.URI); err != nil {
				return err
			}
		}
	}

	return nil
}

// CustomTime represents the custom time format from the API.
type CustomTime struct {
	time.Time
}

// UnmarshalJSON properly unmarshals the time from JSON.
func (t *CustomTime) UnmarshalJSON(b []byte) (err error) {
	t.Time, err = time.Parse(timeFormat, strings.Replace(string(b), "\"", "", -1))

	if err != nil {
		t.Time = time.Time{}
	}

	return err
}
