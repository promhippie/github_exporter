package exporter

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/go-github/v35/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/promhippie/github_exporter/pkg/config"
)

// ActionCollector collects metrics about the servers.
type ActionCollector struct {
	client   *github.Client
	logger   log.Logger
	failures *prometheus.CounterVec
	duration *prometheus.HistogramVec
	config   config.Target

	MinutesUsed          *prometheus.Desc
	MinutesUsedBreakdown *prometheus.Desc
	PaidMinutesUsed      *prometheus.Desc
	IncludedMinutes      *prometheus.Desc
}

// NewActionCollector returns a new ActionCollector.
func NewActionCollector(logger log.Logger, client *github.Client, failures *prometheus.CounterVec, duration *prometheus.HistogramVec, cfg config.Target) *ActionCollector {
	if failures != nil {
		failures.WithLabelValues("action").Add(0)
	}

	labels := []string{"type", "name"}
	return &ActionCollector{
		client:   client,
		logger:   log.With(logger, "collector", "action"),
		failures: failures,
		duration: duration,
		config:   cfg,

		MinutesUsed: prometheus.NewDesc(
			"github_action_billing_minutes_used",
			"Total action minutes used for this type",
			labels,
			nil,
		),
		MinutesUsedBreakdown: prometheus.NewDesc(
			"github_action_billing_minutes_used_breakdown",
			"Total action minutes used for this type broken down by operating system",
			append(labels, "os"),
			nil,
		),
		PaidMinutesUsed: prometheus.NewDesc(
			"github_action_billing_paid_minutes",
			"Total paid minutes used for this type",
			labels,
			nil,
		),
		IncludedMinutes: prometheus.NewDesc(
			"github_action_billing_included_minutes",
			"Included minutes for this type",
			labels,
			nil,
		),
	}
}

// Metrics simply returns the list metric descriptors for generating a documentation.
func (c *ActionCollector) Metrics() []*prometheus.Desc {
	return []*prometheus.Desc{
		c.MinutesUsed,
		c.MinutesUsedBreakdown,
		c.PaidMinutesUsed,
		c.IncludedMinutes,
	}
}

// Describe sends the super-set of all possible descriptors of metrics collected by this Collector.
func (c *ActionCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.MinutesUsed
	ch <- c.MinutesUsedBreakdown
	ch <- c.PaidMinutesUsed
	ch <- c.IncludedMinutes
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *ActionCollector) Collect(ch chan<- prometheus.Metric) {
	for _, name := range c.config.Enterprises.Value() {
		ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
		defer cancel()

		req, err := c.client.NewRequest(
			"GET",
			fmt.Sprintf("/enterprises/%s/settings/billing/actions", name),
			nil,
		)

		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to prepare request",
				"type", "enterprise",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("action").Inc()
			continue
		}

		record := &actionReponse{}
		now := time.Now()
		_, err = c.client.Do(ctx, req, record)
		c.duration.WithLabelValues("action").Observe(time.Since(now).Seconds())

		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to fetch billing",
				"type", "enterprise",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("action").Inc()
			continue
		}

		labels := []string{
			"enterprise",
			name,
		}

		ch <- prometheus.MustNewConstMetric(
			c.MinutesUsed,
			prometheus.GaugeValue,
			float64(record.TotalMinutesUsed),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PaidMinutesUsed,
			prometheus.GaugeValue,
			float64(record.TotalPaidMinutesUsed),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.IncludedMinutes,
			prometheus.GaugeValue,
			float64(record.IncludedMinutes),
			labels...,
		)

		for os, value := range record.MinutesUsedBreakdown {
			ch <- prometheus.MustNewConstMetric(
				c.MinutesUsedBreakdown,
				prometheus.GaugeValue,
				float64(value),
				append(labels, os)...,
			)
		}
	}

	for _, name := range c.config.Orgs.Value() {
		ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
		defer cancel()

		req, err := c.client.NewRequest(
			"GET",
			fmt.Sprintf("/orgs/%s/settings/billing/actions", name),
			nil,
		)

		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to prepare request",
				"type", "org",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("action").Inc()
			continue
		}

		record := &actionReponse{}
		now := time.Now()
		_, err = c.client.Do(ctx, req, record)
		c.duration.WithLabelValues("action").Observe(time.Since(now).Seconds())

		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to fetch billing",
				"type", "org",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("action").Inc()
			continue
		}

		labels := []string{
			"org",
			name,
		}

		ch <- prometheus.MustNewConstMetric(
			c.MinutesUsed,
			prometheus.GaugeValue,
			float64(record.TotalMinutesUsed),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PaidMinutesUsed,
			prometheus.GaugeValue,
			float64(record.TotalPaidMinutesUsed),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.IncludedMinutes,
			prometheus.GaugeValue,
			float64(record.IncludedMinutes),
			labels...,
		)

		for os, value := range record.MinutesUsedBreakdown {
			ch <- prometheus.MustNewConstMetric(
				c.MinutesUsedBreakdown,
				prometheus.GaugeValue,
				float64(value),
				append(labels, os)...,
			)
		}
	}
}

type actionReponse struct {
	TotalMinutesUsed     int            `json:"total_minutes_used"`
	TotalPaidMinutesUsed int            `json:"total_paid_minutes_used"`
	IncludedMinutes      int            `json:"included_minutes"`
	MinutesUsedBreakdown map[string]int `json:"minutes_used_breakdown"`
}
