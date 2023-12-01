package exporter

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/google/go-github/v57/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/promhippie/github_exporter/pkg/config"
	"github.com/promhippie/github_exporter/pkg/store"
)

// BillingCollector collects metrics about the servers.
type BillingCollector struct {
	client   *github.Client
	logger   log.Logger
	db       store.Store
	failures *prometheus.CounterVec
	duration *prometheus.HistogramVec
	config   config.Target

	MinutesUsed          *prometheus.Desc
	MinutesUsedBreakdown *prometheus.Desc
	PaidMinutesUsed      *prometheus.Desc
	IncludedMinutes      *prometheus.Desc

	BandwidthUsed     *prometheus.Desc
	BandwidthPaid     *prometheus.Desc
	BandwidthIncluded *prometheus.Desc

	DaysLeft              *prometheus.Desc
	EastimatedPaidStorage *prometheus.Desc
	EastimatedStorage     *prometheus.Desc
}

// NewBillingCollector returns a new BillingCollector.
func NewBillingCollector(logger log.Logger, client *github.Client, db store.Store, failures *prometheus.CounterVec, duration *prometheus.HistogramVec, cfg config.Target) *BillingCollector {
	if failures != nil {
		failures.WithLabelValues("billing").Add(0)
	}

	labels := []string{"type", "name"}
	return &BillingCollector{
		client:   client,
		logger:   log.With(logger, "collector", "billing"),
		db:       db,
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
func (c *BillingCollector) Metrics() []*prometheus.Desc {
	return []*prometheus.Desc{
		c.MinutesUsed,
		c.MinutesUsedBreakdown,
		c.PaidMinutesUsed,
		c.IncludedMinutes,
		c.BandwidthUsed,
		c.BandwidthPaid,
		c.BandwidthIncluded,
		c.DaysLeft,
		c.EastimatedPaidStorage,
		c.EastimatedStorage,
	}
}

// Describe sends the super-set of all possible descriptors of metrics collected by this Collector.
func (c *BillingCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.MinutesUsed
	ch <- c.MinutesUsedBreakdown
	ch <- c.PaidMinutesUsed
	ch <- c.IncludedMinutes
	ch <- c.BandwidthUsed
	ch <- c.BandwidthPaid
	ch <- c.BandwidthIncluded
	ch <- c.DaysLeft
	ch <- c.EastimatedPaidStorage
	ch <- c.EastimatedStorage
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *BillingCollector) Collect(ch chan<- prometheus.Metric) {
	{
		collected := make([]string, 0)

		now := time.Now()
		billing := c.getActionBilling()
		c.duration.WithLabelValues("action").Observe(time.Since(now).Seconds())

		level.Debug(c.logger).Log(
			"msg", "Fetched action billing",
			"count", len(billing),
			"duration", time.Since(now),
		)

		for _, record := range billing {
			if alreadyCollected(collected, record.Name) {
				level.Debug(c.logger).Log(
					"msg", "Already collected action billing",
					"type", record.Type,
					"name", record.Name,
				)

				continue
			}

			collected = append(collected, record.Name)

			level.Debug(c.logger).Log(
				"msg", "Collecting action billing",
				"type", record.Type,
				"name", record.Name,
			)

			labels := []string{
				record.Type,
				record.Name,
			}

			ch <- prometheus.MustNewConstMetric(
				c.MinutesUsed,
				prometheus.GaugeValue,
				record.TotalMinutesUsed,
				labels...,
			)

			ch <- prometheus.MustNewConstMetric(
				c.PaidMinutesUsed,
				prometheus.GaugeValue,
				record.TotalPaidMinutesUsed,
				labels...,
			)

			ch <- prometheus.MustNewConstMetric(
				c.IncludedMinutes,
				prometheus.GaugeValue,
				record.IncludedMinutes,
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

	{
		collected := make([]string, 0)

		now := time.Now()
		billing := c.getPackageBilling()
		c.duration.WithLabelValues("action").Observe(time.Since(now).Seconds())

		level.Debug(c.logger).Log(
			"msg", "Fetched package billing",
			"count", len(billing),
			"duration", time.Since(now),
		)

		for _, record := range billing {
			if alreadyCollected(collected, record.Name) {
				level.Debug(c.logger).Log(
					"msg", "Already collected package billing",
					"type", record.Type,
					"name", record.Name,
				)

				continue
			}

			collected = append(collected, record.Name)

			level.Debug(c.logger).Log(
				"msg", "Collecting package billing",
				"type", record.Type,
				"name", record.Name,
			)

			labels := []string{
				record.Type,
				record.Name,
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
				record.IncludedGigabytesBandwidth,
				labels...,
			)
		}
	}

	{
		collected := make([]string, 0)

		now := time.Now()
		billing := c.getStorageBilling()
		c.duration.WithLabelValues("action").Observe(time.Since(now).Seconds())

		level.Debug(c.logger).Log(
			"msg", "Fetched storage billing",
			"count", len(billing),
			"duration", time.Since(now),
		)

		for _, record := range billing {
			if alreadyCollected(collected, record.Name) {
				level.Debug(c.logger).Log(
					"msg", "Already collected storage billing",
					"type", record.Type,
					"name", record.Name,
				)

				continue
			}

			collected = append(collected, record.Name)

			level.Debug(c.logger).Log(
				"msg", "Collecting storage billing",
				"type", record.Type,
				"name", record.Name,
			)

			labels := []string{
				record.Type,
				record.Name,
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
				record.EstimatedPaidStorageForMonth,
				labels...,
			)

			ch <- prometheus.MustNewConstMetric(
				c.EastimatedStorage,
				prometheus.GaugeValue,
				record.EstimatedStorageForMonth,
				labels...,
			)
		}
	}
}

type actionBilling struct {
	Type string
	Name string
	*github.ActionBilling
}

func (c *BillingCollector) getActionBilling() []*actionBilling {
	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	result := make([]*actionBilling, 0)

	for _, name := range c.config.Enterprises.Value() {
		req, err := c.client.NewRequest(
			"GET",
			fmt.Sprintf("/enterprises/%s/settings/billing/actions", name),
			nil,
		)

		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to prepare action request",
				"type", "enterprise",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("action").Inc()
			continue
		}

		record := &github.ActionBilling{}
		resp, err := c.client.Do(ctx, req, record)

		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to fetch action billing",
				"type", "enterprise",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("action").Inc()
			continue
		}

		defer closeBody(resp)

		result = append(result, &actionBilling{
			Type:          "enterprise",
			Name:          name,
			ActionBilling: record,
		})
	}

	for _, name := range c.config.Orgs.Value() {
		record, resp, err := c.client.Billing.GetActionsBillingOrg(ctx, name)
		defer closeBody(resp)

		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to fetch action billing",
				"type", "org",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("action").Inc()
			continue
		}

		result = append(result, &actionBilling{
			Type:          "org",
			Name:          name,
			ActionBilling: record,
		})
	}

	return result
}

type packageBilling struct {
	Type string
	Name string
	*github.PackageBilling
}

func (c *BillingCollector) getPackageBilling() []*packageBilling {
	result := make([]*packageBilling, 0)

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
				"msg", "Failed to prepare package request",
				"type", "enterprise",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("action").Inc()
			continue
		}

		record := &github.PackageBilling{}
		resp, err := c.client.Do(ctx, req, record)

		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to fetch package billing",
				"type", "enterprise",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("action").Inc()
			continue
		}

		defer closeBody(resp)

		result = append(result, &packageBilling{
			Type:           "enterprise",
			Name:           name,
			PackageBilling: record,
		})
	}

	for _, name := range c.config.Orgs.Value() {
		ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
		defer cancel()

		record, resp, err := c.client.Billing.GetPackagesBillingOrg(ctx, name)
		defer closeBody(resp)

		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to fetch package billing",
				"type", "org",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("action").Inc()
			continue
		}

		result = append(result, &packageBilling{
			Type:           "org",
			Name:           name,
			PackageBilling: record,
		})
	}

	return result
}

type storageBilling struct {
	Type string
	Name string
	*github.StorageBilling
}

func (c *BillingCollector) getStorageBilling() []*storageBilling {
	result := make([]*storageBilling, 0)

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
				"msg", "Failed to prepare storage request",
				"type", "enterprise",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("action").Inc()
			continue
		}

		record := &github.StorageBilling{}
		resp, err := c.client.Do(ctx, req, record)

		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to fetch storage billing",
				"type", "enterprise",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("action").Inc()
			continue
		}

		defer closeBody(resp)

		result = append(result, &storageBilling{
			Type:           "enterprise",
			Name:           name,
			StorageBilling: record,
		})
	}

	for _, name := range c.config.Orgs.Value() {
		ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
		defer cancel()

		record, resp, err := c.client.Billing.GetStorageBillingOrg(ctx, name)
		defer closeBody(resp)

		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to fetch storage billing",
				"type", "org",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("action").Inc()
			continue
		}

		result = append(result, &storageBilling{
			Type:           "org",
			Name:           name,
			StorageBilling: record,
		})
	}

	return result
}
