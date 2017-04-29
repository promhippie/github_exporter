package exporter

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const (
	// namespace defines the Prometheus namespace for this exporter.
	namespace = "github"
)

var (
	// validResponse defines if the API response can get processed.
	validResponse = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "valid_response",
			Help:      "Check if GitHub response can be processed",
		},
	)

	// openIssues defines a map to collect the open issues per repository.
	openIssues = map[string]prometheus.Gauge{}

	// forkCount defines a map to collect the number of forks per repository.
	forkCount = map[string]prometheus.Gauge{}

	// starCount defines a map to collect the number of stars per repository.
	starCount = map[string]prometheus.Gauge{}

	// watchCount defines a map to collect the number of watchers per repository.
	watchCount = map[string]prometheus.Gauge{}

	// sizeValue defines a map to collect the size of the repositories.
	sizeValue = map[string]prometheus.Gauge{}

	// pushedAt defines a map to collect the last push timestamp per repository.
	pushedAt = map[string]prometheus.Gauge{}

	// updatedAt defines a map to collect the last update timestamp per repository.
	updatedAt = map[string]prometheus.Gauge{}
)

// init just defines the initial state of the exports.
func init() {
	validResponse.Set(0)
}

// NewExporter gives you a new exporter instance.
func NewExporter(orgs, repos []string) *Exporter {
	return &Exporter{
		orgs:  orgs,
		repos: repos,
	}
}

// Exporter combines the metric collector and descritions.
type Exporter struct {
	orgs  []string
	repos []string
	mutex sync.RWMutex
}

// Describe defines the metric descriptions for Prometheus.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- validResponse.Desc()

	for _, metric := range openIssues {
		ch <- metric.Desc()
	}

	for _, metric := range forkCount {
		ch <- metric.Desc()
	}

	for _, metric := range starCount {
		ch <- metric.Desc()
	}

	for _, metric := range watchCount {
		ch <- metric.Desc()
	}

	for _, metric := range sizeValue {
		ch <- metric.Desc()
	}

	for _, metric := range pushedAt {
		ch <- metric.Desc()
	}

	for _, metric := range updatedAt {
		ch <- metric.Desc()
	}
}

// Collect delivers the metrics to Prometheus.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if err := e.scrape(); err != nil {
		log.Error(err)

		validResponse.Set(0)
		ch <- validResponse

		return
	}

	ch <- validResponse

	for _, metric := range openIssues {
		ch <- metric
	}

	for _, metric := range forkCount {
		ch <- metric
	}

	for _, metric := range starCount {
		ch <- metric
	}

	for _, metric := range watchCount {
		ch <- metric
	}

	for _, metric := range sizeValue {
		ch <- metric
	}

	for _, metric := range pushedAt {
		ch <- metric
	}

	for _, metric := range updatedAt {
		ch <- metric
	}
}

// scrape just starts the scraping loop.
func (e *Exporter) scrape() error {
	log.Debug("start scrape loop")

	for _, org := range e.orgs {
		log.Debugf("checking %s organization", org)

		if err := processOrg(org); err != nil {
			return err
		}
	}

	for _, repo := range e.repos {
		log.Debugf("checking %s repository", repo)

		if err := processRepo(repo); err != nil {
			return err
		}
	}

	validResponse.Set(1)
	return nil
}

// processOrg fetches the organization content from the API.
func processOrg(name string) error {
	var (
		collection = &Collection{}
	)

	if err := collection.Fetch(name); err != nil {
		log.Debugf("%s", err)
		return fmt.Errorf("failed to fetch %s organization", name)
	}

	for _, repo := range collection.Repos {
		if err := scrapeRepo(repo); err != nil {
			return err
		}
	}

	return nil
}

// processRepo fetches the repository content from the API.
func processRepo(name string) error {
	var (
		repo = &Repo{}
	)

	if err := repo.Fetch(name); err != nil {
		log.Debugf("%s", err)
		return fmt.Errorf("failed to fetch %s repository", name)
	}

	return scrapeRepo(repo)
}

// scrapeRepo processes the content of a specific repository.
func scrapeRepo(repo *Repo) error {
	if _, ok := openIssues[repo.Key()]; ok == false {
		openIssues[repo.Key()] = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "issues",
				Help:      "How many open issues does the repository have",
				ConstLabels: prometheus.Labels{
					"owner": repo.Owner.Login,
					"repo":  repo.Name,
				},
			},
		)
	}

	if _, ok := forkCount[repo.Key()]; ok == false {
		forkCount[repo.Key()] = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "forks",
				Help:      "How often have this repository been forked",
				ConstLabels: prometheus.Labels{
					"owner": repo.Owner.Login,
					"repo":  repo.Name,
				},
			},
		)
	}

	if _, ok := starCount[repo.Key()]; ok == false {
		starCount[repo.Key()] = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "stars",
				Help:      "How often have this repository been stared",
				ConstLabels: prometheus.Labels{
					"owner": repo.Owner.Login,
					"repo":  repo.Name,
				},
			},
		)
	}

	if _, ok := watchCount[repo.Key()]; ok == false {
		watchCount[repo.Key()] = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "watchers",
				Help:      "How often have this repository been watched",
				ConstLabels: prometheus.Labels{
					"owner": repo.Owner.Login,
					"repo":  repo.Name,
				},
			},
		)
	}

	if _, ok := sizeValue[repo.Key()]; ok == false {
		sizeValue[repo.Key()] = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "size",
				Help:      "Simply the size of the Git repository",
				ConstLabels: prometheus.Labels{
					"owner": repo.Owner.Login,
					"repo":  repo.Name,
				},
			},
		)
	}

	if _, ok := pushedAt[repo.Key()]; ok == false {
		pushedAt[repo.Key()] = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "pushed",
				Help:      "A timestamp when the repository had the last push",
				ConstLabels: prometheus.Labels{
					"owner": repo.Owner.Login,
					"repo":  repo.Name,
				},
			},
		)
	}

	if _, ok := updatedAt[repo.Key()]; ok == false {
		updatedAt[repo.Key()] = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "updated",
				Help:      "A timestamp when the repository have been updated",
				ConstLabels: prometheus.Labels{
					"owner": repo.Owner.Login,
					"repo":  repo.Name,
				},
			},
		)
	}

	openIssues[repo.Key()].Set(repo.Issues)
	forkCount[repo.Key()].Set(repo.Forks)
	starCount[repo.Key()].Set(repo.Stars)
	watchCount[repo.Key()].Set(repo.Watchers)
	sizeValue[repo.Key()].Set(repo.Size)
	pushedAt[repo.Key()].Set(float64(repo.PushedAt.Unix()))
	updatedAt[repo.Key()].Set(float64(repo.UpdatedAt.Unix()))

	return nil
}
