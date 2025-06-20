package exporter

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/go-github/v72/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/promhippie/github_exporter/pkg/config"
	"github.com/promhippie/github_exporter/pkg/store"
)

// BillingCollector collects metrics about the servers.
type BillingCollector struct {
	client   *github.Client
	logger   *slog.Logger
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
func NewBillingCollector(logger *slog.Logger, client *github.Client, db store.Store, failures *prometheus.CounterVec, duration *prometheus.HistogramVec, cfg config.Target) *BillingCollector {
	if failures != nil {
		failures.WithLabelValues("billing").Add(0)
	}

	labels := []string{"type", "name"}
	return &BillingCollector{
		client:   client,
		logger:   logger.With("collector", "billing"),
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

		c.logger.Debug("Fetched action billing",
			"count", len(billing),
			"duration", time.Since(now),
		)

		for _, record := range billing {
			if alreadyCollected(collected, record.Name) {
				c.logger.Debug("Already collected action billing",
					"type", record.Type,
					"name", record.Name,
				)

				continue
			}

			collected = append(collected, record.Name)

			c.logger.Debug("Collecting action billing",
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

		c.logger.Debug("Fetched package billing",
			"count", len(billing),
			"duration", time.Since(now),
		)

		for _, record := range billing {
			if alreadyCollected(collected, record.Name) {
				c.logger.Debug("Already collected package billing",
					"type", record.Type,
					"name", record.Name,
				)

				continue
			}

			collected = append(collected, record.Name)

			c.logger.Debug("Collecting package billing",
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

		c.logger.Debug("Fetched storage billing",
			"count", len(billing),
			"duration", time.Since(now),
		)

		for _, record := range billing {
			if alreadyCollected(collected, record.Name) {
				c.logger.Debug("Already collected storage billing",
					"type", record.Type,
					"name", record.Name,
				)

				continue
			}

			collected = append(collected, record.Name)

			c.logger.Debug("Collecting storage billing",
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

	for _, name := range c.config.Enterprises {
		req, err := c.client.NewRequest(
			"GET",
			fmt.Sprintf("/enterprises/%s/settings/billing/usage", name),
			nil,
		)

		if err != nil {
			c.logger.Error("Failed to prepare action request",
				"type", "enterprise",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("action").Inc()
			continue
		}

		usage := &UnifiedBillingResponse{}
		resp, err := c.client.Do(ctx, req, usage)

		if err != nil {
			c.logger.Error("Failed to fetch action billing",
				"type", "enterprise",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("action").Inc()
			continue
		}

		defer closeBody(resp)

		record := transformToActionBilling(usage)
		result = append(result, &actionBilling{
			Type:          "enterprise",
			Name:          name,
			ActionBilling: record,
		})
	}

	for _, name := range c.config.Orgs {
		req, err := c.client.NewRequest(
			"GET",
			fmt.Sprintf("/organizations/%s/settings/billing/usage", name),
			nil,
		)

		if err != nil {
			c.logger.Error("Failed to prepare action request",
				"type", "org",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("action").Inc()
			continue
		}

		usage := &UnifiedBillingResponse{}
		resp, err := c.client.Do(ctx, req, usage)

		if err != nil {
			c.logger.Error("Failed to fetch action billing",
				"type", "org",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("action").Inc()
			continue
		}

		defer closeBody(resp)

		record := transformToActionBilling(usage)
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
	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	result := make([]*packageBilling, 0)

	for _, name := range c.config.Enterprises {
		req, err := c.client.NewRequest(
			"GET",
			fmt.Sprintf("/enterprises/%s/settings/billing/usage", name),
			nil,
		)

		if err != nil {
			c.logger.Error("Failed to prepare package request",
				"type", "enterprise",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("action").Inc()
			continue
		}

		usage := &UnifiedBillingResponse{}
		resp, err := c.client.Do(ctx, req, usage)

		if err != nil {
			c.logger.Error("Failed to fetch package billing",
				"type", "enterprise",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("action").Inc()
			continue
		}

		defer closeBody(resp)

		record := transformToPackageBilling(usage)
		result = append(result, &packageBilling{
			Type:           "enterprise",
			Name:           name,
			PackageBilling: record,
		})
	}

	for _, name := range c.config.Orgs {
		req, err := c.client.NewRequest(
			"GET",
			fmt.Sprintf("/organizations/%s/settings/billing/usage", name),
			nil,
		)

		if err != nil {
			c.logger.Error("Failed to prepare package request",
				"type", "org",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("action").Inc()
			continue
		}

		usage := &UnifiedBillingResponse{}
		resp, err := c.client.Do(ctx, req, usage)

		if err != nil {
			c.logger.Error("Failed to fetch package billing",
				"type", "org",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("action").Inc()
			continue
		}

		defer closeBody(resp)

		record := transformToPackageBilling(usage)
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

// UnifiedBillingResponse represents the new unified billing API response
type UnifiedBillingResponse struct {
	UsageItems []UsageItem `json:"usageItems"`
}

// UsageItem represents an individual usage item in the unified billing response
type UsageItem struct {
	Date             string  `json:"date"`
	Product          string  `json:"product"`
	SKU              string  `json:"sku"`
	Quantity         int     `json:"quantity"`
	UnitType         string  `json:"unitType"`
	PricePerUnit     float64 `json:"pricePerUnit"`
	GrossAmount      float64 `json:"grossAmount"`
	DiscountAmount   float64 `json:"discountAmount"`
	NetAmount        float64 `json:"netAmount"`
	OrganizationName string  `json:"organizationName"`
	RepositoryName   string  `json:"repositoryName"`
}

// transformToActionBilling converts unified billing data to legacy ActionBilling format
func transformToActionBilling(usage *UnifiedBillingResponse) *github.ActionBilling {
	var totalMinutesUsed, totalPaidMinutesUsed, includedMinutes float64
	minutesBreakdown := make(map[string]int)

	for _, item := range usage.UsageItems {
		if item.Product == "Actions" && item.UnitType == "minutes" {
			quantity := float64(item.Quantity)
			totalMinutesUsed += quantity

			// Calculate paid minutes from net amount (dollar value)
			if item.NetAmount > 0 {
				totalPaidMinutesUsed += item.NetAmount
			}

			// Calculate included minutes from discount amount
			if item.DiscountAmount > 0 && item.PricePerUnit > 0 {
				includedMinutes += item.DiscountAmount / item.PricePerUnit
			}

			// Map SKU to OS for breakdown
			var os string
			if strings.Contains(strings.ToLower(item.SKU), "linux") {
				os = "UBUNTU"
			} else if strings.Contains(strings.ToLower(item.SKU), "windows") {
				os = "WINDOWS"
			} else if strings.Contains(strings.ToLower(item.SKU), "macos") {
				os = "MACOS"
			}

			if os != "" {
				minutesBreakdown[os] += item.Quantity
			}
		}
	}

	return &github.ActionBilling{
		TotalMinutesUsed:     totalMinutesUsed,
		TotalPaidMinutesUsed: totalPaidMinutesUsed,
		IncludedMinutes:      includedMinutes,
		MinutesUsedBreakdown: minutesBreakdown,
	}
}

// transformToPackageBilling converts unified billing data to legacy PackageBilling format
func transformToPackageBilling(usage *UnifiedBillingResponse) *github.PackageBilling {
	var totalBandwidthUsed, totalPaidBandwidthUsed, includedBandwidth int

	for _, item := range usage.UsageItems {
		if item.Product == "Packages" && strings.Contains(strings.ToLower(item.SKU), "data transfer") {
			totalBandwidthUsed += item.Quantity

			// Count as paid if there's a net amount
			if item.NetAmount > 0 {
				totalPaidBandwidthUsed += item.Quantity
			}

			// Calculate included bandwidth from discount amount
			if item.DiscountAmount > 0 && item.PricePerUnit > 0 {
				includedBandwidth += int(item.DiscountAmount / item.PricePerUnit)
			}
		}
	}

	return &github.PackageBilling{
		TotalGigabytesBandwidthUsed:     totalBandwidthUsed,
		TotalPaidGigabytesBandwidthUsed: totalPaidBandwidthUsed,
		IncludedGigabytesBandwidth:      float64(includedBandwidth),
	}
}

// transformToStorageBilling converts unified billing data to legacy StorageBilling format
func transformToStorageBilling(usage *UnifiedBillingResponse) *github.StorageBilling {
	var estimatedPaidStorage, estimatedStorage float64

	for _, item := range usage.UsageItems {
		if strings.Contains(strings.ToLower(item.SKU), "storage") {
			estimatedStorage += item.GrossAmount
			if item.NetAmount > 0 {
				estimatedPaidStorage += item.NetAmount
			}
		}
	}

	// Calculate days left in billing cycle (approximate)
	now := time.Now()
	year, month, _ := now.Date()
	nextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, now.Location())
	daysLeft := int(nextMonth.Sub(now).Hours() / 24)

	return &github.StorageBilling{
		DaysLeftInBillingCycle:        daysLeft,
		EstimatedPaidStorageForMonth:  estimatedPaidStorage,
		EstimatedStorageForMonth:      estimatedStorage,
	}
}

func (c *BillingCollector) getStorageBilling() []*storageBilling {
	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	result := make([]*storageBilling, 0)

	for _, name := range c.config.Enterprises {
		req, err := c.client.NewRequest(
			"GET",
			fmt.Sprintf("/enterprises/%s/settings/billing/usage", name),
			nil,
		)

		if err != nil {
			c.logger.Error("Failed to prepare storage request",
				"type", "enterprise",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("action").Inc()
			continue
		}

		usage := &UnifiedBillingResponse{}
		resp, err := c.client.Do(ctx, req, usage)

		if err != nil {
			c.logger.Error("Failed to fetch storage billing",
				"type", "enterprise",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("action").Inc()
			continue
		}

		defer closeBody(resp)

		record := transformToStorageBilling(usage)
		result = append(result, &storageBilling{
			Type:           "enterprise",
			Name:           name,
			StorageBilling: record,
		})
	}

	for _, name := range c.config.Orgs {
		req, err := c.client.NewRequest(
			"GET",
			fmt.Sprintf("/organizations/%s/settings/billing/usage", name),
			nil,
		)

		if err != nil {
			c.logger.Error("Failed to prepare storage request",
				"type", "org",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("action").Inc()
			continue
		}

		usage := &UnifiedBillingResponse{}
		resp, err := c.client.Do(ctx, req, usage)

		if err != nil {
			c.logger.Error("Failed to fetch storage billing",
				"type", "org",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("action").Inc()
			continue
		}

		defer closeBody(resp)

		record := transformToStorageBilling(usage)
		result = append(result, &storageBilling{
			Type:           "org",
			Name:           name,
			StorageBilling: record,
		})
	}

	return result
}
