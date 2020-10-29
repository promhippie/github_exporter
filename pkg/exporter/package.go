package exporter

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/go-github/v32/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/promhippie/github_exporter/pkg/config"
)

// PackageCollector collects metrics about the servers.
type PackageCollector struct {
	client   *github.Client
	logger   log.Logger
	failures *prometheus.CounterVec
	duration *prometheus.HistogramVec
	config   config.Target

	BandwidthUsed     *prometheus.Desc
	BandwidthPaid     *prometheus.Desc
	BandwidthIncluded *prometheus.Desc
}

// NewPackageCollector returns a new PackageCollector.
func NewPackageCollector(logger log.Logger, client *github.Client, failures *prometheus.CounterVec, duration *prometheus.HistogramVec, cfg config.Target) *PackageCollector {
	failures.WithLabelValues("package").Add(0)

	labels := []string{"type", "name"}
	return &PackageCollector{
		client:   client,
		logger:   log.With(logger, "collector", "package"),
		failures: failures,
		duration: duration,
		config:   cfg,

		BandwidthUsed: prometheus.NewDesc(
			"github_package_billing_gigabytes_bandwidth_used",
			"Total bandwidth used by this type in Gigabytes",
			labels,
			nil,
		),
		BandwidthPaid: prometheus.NewDesc(
			"github_package_billing_paid_gigabytes_bandwidth_used",
			"Total paid bandwidth used by this type in Gigabytes",
			labels,
			nil,
		),
		BandwidthIncluded: prometheus.NewDesc(
			"github_package_billing_included_gigabytes_bandwidth",
			"Included bandwidth for this type in Gigabytes",
			labels,
			nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics collected by this Collector.
func (c *PackageCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.BandwidthUsed
	ch <- c.BandwidthPaid
	ch <- c.BandwidthIncluded
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *PackageCollector) Collect(ch chan<- prometheus.Metric) {
	for _, name := range c.config.Enterprises.Value() {
		ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
		defer cancel()

		req, err := c.client.NewRequest(
			"GET",
			fmt.Sprintf("/enterprises/%s/settings/billing/packages", name),
			nil,
		)

		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to prepare request",
				"type", "enterprise",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("package").Inc()
			continue
		}

		record := &packageReponse{}
		now := time.Now()
		_, err = c.client.Do(ctx, req, record)
		c.duration.WithLabelValues("package").Observe(time.Since(now).Seconds())

		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to fetch billing",
				"type", "enterprise",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("package").Inc()
			continue
		}

		labels := []string{
			"enterprise",
			name,
		}

		ch <- prometheus.MustNewConstMetric(
			c.BandwidthUsed,
			prometheus.GaugeValue,
			float64(record.TotalGigabytesBandwidthUsed),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BandwidthPaid,
			prometheus.GaugeValue,
			float64(record.TotalPaidGigabytesBandwidthUsed),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BandwidthIncluded,
			prometheus.GaugeValue,
			float64(record.IncludedGigabytesBandwidth),
			labels...,
		)
	}

	for _, name := range c.config.Orgs.Value() {
		ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
		defer cancel()

		req, err := c.client.NewRequest(
			"GET",
			fmt.Sprintf("/orgs/%s/settings/billing/packages", name),
			nil,
		)

		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to prepare request",
				"type", "org",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("package").Inc()
			continue
		}

		record := &packageReponse{}
		now := time.Now()
		_, err = c.client.Do(ctx, req, record)
		c.duration.WithLabelValues("package").Observe(time.Since(now).Seconds())

		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to fetch billing",
				"type", "org",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("package").Inc()
			continue
		}

		labels := []string{
			"org",
			name,
		}

		ch <- prometheus.MustNewConstMetric(
			c.BandwidthUsed,
			prometheus.GaugeValue,
			float64(record.TotalGigabytesBandwidthUsed),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BandwidthPaid,
			prometheus.GaugeValue,
			float64(record.TotalPaidGigabytesBandwidthUsed),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BandwidthIncluded,
			prometheus.GaugeValue,
			float64(record.IncludedGigabytesBandwidth),
			labels...,
		)
	}
}

type packageReponse struct {
	TotalGigabytesBandwidthUsed     int `json:"total_gigabytes_bandwidth_used"`
	TotalPaidGigabytesBandwidthUsed int `json:"total_paid_gigabytes_bandwidth_used"`
	IncludedGigabytesBandwidth      int `json:"included_gigabytes_bandwidth"`
}
