package exporter

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/google/go-github/v50/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/promhippie/github_exporter/pkg/config"
)

// StorageCollector collects metrics about the servers.
type StorageCollector struct {
	client   *github.Client
	logger   log.Logger
	failures *prometheus.CounterVec
	duration *prometheus.HistogramVec
	config   config.Target

	DaysLeft              *prometheus.Desc
	EastimatedPaidStorage *prometheus.Desc
	EastimatedStorage     *prometheus.Desc
}

// NewStorageCollector returns a new StorageCollector.
func NewStorageCollector(logger log.Logger, client *github.Client, failures *prometheus.CounterVec, duration *prometheus.HistogramVec, cfg config.Target) *StorageCollector {
	if failures != nil {
		failures.WithLabelValues("storage").Add(0)
	}

	labels := []string{"type", "name"}
	return &StorageCollector{
		client:   client,
		logger:   log.With(logger, "collector", "storage"),
		failures: failures,
		duration: duration,
		config:   cfg,

		DaysLeft: prometheus.NewDesc(
			"github_storage_billing_days_left_in_cycle",
			"Days left within this billing cycle for this type",
			labels,
			nil,
		),
		EastimatedPaidStorage: prometheus.NewDesc(
			"github_storage_billing_estimated_paid_storage_for_month",
			"Estimated paid storage for this month for this type",
			labels,
			nil,
		),
		EastimatedStorage: prometheus.NewDesc(
			"github_storage_billing_estimated_storage_for_month",
			"Estimated total storage for this month for this type",
			labels,
			nil,
		),
	}
}

// Metrics simply returns the list metric descriptors for generating a documentation.
func (c *StorageCollector) Metrics() []*prometheus.Desc {
	return []*prometheus.Desc{
		c.DaysLeft,
		c.EastimatedPaidStorage,
		c.EastimatedStorage,
	}
}

// Describe sends the super-set of all possible descriptors of metrics collected by this Collector.
func (c *StorageCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.DaysLeft
	ch <- c.EastimatedPaidStorage
	ch <- c.EastimatedStorage
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *StorageCollector) Collect(ch chan<- prometheus.Metric) {
	for _, name := range c.config.Enterprises.Value() {
		ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
		defer cancel()

		req, err := c.client.NewRequest(
			"GET",
			fmt.Sprintf("/enterprises/%s/settings/billing/shared-storage", name),
			nil,
		)

		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to prepare request",
				"type", "enterprise",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("storage").Inc()
			continue
		}

		record := &storageReponse{}
		now := time.Now()
		_, err = c.client.Do(ctx, req, record)
		c.duration.WithLabelValues("storage").Observe(time.Since(now).Seconds())

		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to fetch billing",
				"type", "enterprise",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("storage").Inc()
			continue
		}

		labels := []string{
			"enterprise",
			name,
		}

		ch <- prometheus.MustNewConstMetric(
			c.DaysLeft,
			prometheus.GaugeValue,
			float64(record.DaysLeftInBillingCycle),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.EastimatedPaidStorage,
			prometheus.GaugeValue,
			float64(record.EstimatedPaidStorageForMonth),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.EastimatedStorage,
			prometheus.GaugeValue,
			float64(record.EstimatedStorageForMonth),
			labels...,
		)
	}

	for _, name := range c.config.Orgs.Value() {
		ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
		defer cancel()

		req, err := c.client.NewRequest(
			"GET",
			fmt.Sprintf("/orgs/%s/settings/billing/shared-storage", name),
			nil,
		)

		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to prepare request",
				"type", "org",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("storage").Inc()
			continue
		}

		record := &storageReponse{}
		now := time.Now()
		_, err = c.client.Do(ctx, req, record)
		c.duration.WithLabelValues("storage").Observe(time.Since(now).Seconds())

		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to fetch billing",
				"type", "org",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("storage").Inc()
			continue
		}

		labels := []string{
			"org",
			name,
		}

		ch <- prometheus.MustNewConstMetric(
			c.DaysLeft,
			prometheus.GaugeValue,
			float64(record.DaysLeftInBillingCycle),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.EastimatedPaidStorage,
			prometheus.GaugeValue,
			float64(record.EstimatedPaidStorageForMonth),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.EastimatedStorage,
			prometheus.GaugeValue,
			float64(record.EstimatedStorageForMonth),
			labels...,
		)
	}
}

type storageReponse struct {
	DaysLeftInBillingCycle       int `json:"days_left_in_billing_cycle"`
	EstimatedPaidStorageForMonth int `json:"estimated_paid_storage_for_month"`
	EstimatedStorageForMonth     int `json:"estimated_storage_for_month"`
}
