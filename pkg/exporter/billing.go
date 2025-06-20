package exporter

import (
	"context"
	"fmt"
	"log/slog"
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

	// Enhanced Billing Platform metrics - v5.0.0 breaking change
	BillingUsage          *prometheus.Desc
	BillingGrossAmount    *prometheus.Desc
	BillingDiscountAmount *prometheus.Desc
	BillingNetAmount      *prometheus.Desc
	BillingPricePerUnit   *prometheus.Desc
}

// NewBillingCollector returns a new BillingCollector.
func NewBillingCollector(logger *slog.Logger, client *github.Client, db store.Store, failures *prometheus.CounterVec, duration *prometheus.HistogramVec, cfg config.Target) *BillingCollector {
	if failures != nil {
		failures.WithLabelValues("billing").Add(0)
	}

	// Enhanced Billing Platform labels - v5.0.0
	labels := []string{"type", "name", "product", "sku", "unit_type", "date", "organization_name", "repository_name"}
	return &BillingCollector{
		client:   client,
		logger:   logger.With("collector", "billing"),
		db:       db,
		failures: failures,
		duration: duration,
		config:   cfg,

		BillingUsage: prometheus.NewDesc(
			"github_billing_usage",
			"Usage quantity from GitHub Enhanced Billing Platform",
			labels,
			nil,
		),
		BillingGrossAmount: prometheus.NewDesc(
			"github_billing_usage_gross_amount",
			"Gross amount charged for this usage item",
			labels,
			nil,
		),
		BillingDiscountAmount: prometheus.NewDesc(
			"github_billing_usage_discount_amount",
			"Discount amount applied to this usage item",
			labels,
			nil,
		),
		BillingNetAmount: prometheus.NewDesc(
			"github_billing_usage_net_amount",
			"Net amount charged for this usage item after discounts",
			labels,
			nil,
		),
		BillingPricePerUnit: prometheus.NewDesc(
			"github_billing_usage_price_per_unit",
			"Price per unit for this usage item",
			labels,
			nil,
		),
	}
}

// Metrics simply returns the list metric descriptors for generating a documentation.
func (c *BillingCollector) Metrics() []*prometheus.Desc {
	return []*prometheus.Desc{
		c.BillingUsage,
		c.BillingGrossAmount,
		c.BillingDiscountAmount,
		c.BillingNetAmount,
		c.BillingPricePerUnit,
	}
}

// Describe sends the super-set of all possible descriptors of metrics collected by this Collector.
func (c *BillingCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.BillingUsage
	ch <- c.BillingGrossAmount
	ch <- c.BillingDiscountAmount
	ch <- c.BillingNetAmount
	ch <- c.BillingPricePerUnit
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *BillingCollector) Collect(ch chan<- prometheus.Metric) {
	now := time.Now()
	usageItems := c.getBillingUsage()
	c.duration.WithLabelValues("billing").Observe(time.Since(now).Seconds())

	c.logger.Debug("Fetched billing usage",
		"count", len(usageItems),
		"duration", time.Since(now),
	)

	collected := make(map[string]bool) // Use to avoid duplicates

	for _, item := range usageItems {
		// Create a unique key for deduplication
		key := fmt.Sprintf("%s-%s-%s-%s-%s", item.Type, item.Name, item.Product, item.SKU, item.Date)

		if collected[key] {
			c.logger.Debug("Already collected billing usage",
				"type", item.Type,
				"name", item.Name,
				"product", item.Product,
				"sku", item.SKU,
				"date", item.Date,
			)
			continue
		}

		collected[key] = true

		c.logger.Debug("Collecting billing usage",
			"type", item.Type,
			"name", item.Name,
			"product", item.Product,
			"sku", item.SKU,
			"unit_type", item.UnitType,
			"quantity", item.Quantity,
		)

		labels := []string{
			item.Type,
			item.Name,
			item.Product,
			item.SKU,
			item.UnitType,
			item.Date,
			item.OrganizationName,
			item.RepositoryName,
		}

		ch <- prometheus.MustNewConstMetric(
			c.BillingUsage,
			prometheus.GaugeValue,
			item.Quantity,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BillingGrossAmount,
			prometheus.GaugeValue,
			item.GrossAmount,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BillingDiscountAmount,
			prometheus.GaugeValue,
			item.DiscountAmount,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BillingNetAmount,
			prometheus.GaugeValue,
			item.NetAmount,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BillingPricePerUnit,
			prometheus.GaugeValue,
			item.PricePerUnit,
			labels...,
		)
	}
}

// UsageItem represents a billing usage item from GitHub Enhanced Billing Platform API
type UsageItem struct {
	Date             string  `json:"date"`
	Product          string  `json:"product"`
	SKU              string  `json:"sku"`
	Quantity         float64 `json:"quantity"`
	UnitType         string  `json:"unitType"`
	PricePerUnit     float64 `json:"pricePerUnit"`
	GrossAmount      float64 `json:"grossAmount"`
	DiscountAmount   float64 `json:"discountAmount"`
	NetAmount        float64 `json:"netAmount"`
	OrganizationName string  `json:"organizationName"`
	RepositoryName   string  `json:"repositoryName"`
	// Additional fields for metrics labeling
	Type string // "org" or "enterprise"
	Name string // organization or enterprise name
}

// UsageResponse represents the response from GitHub Enhanced Billing Platform API
type UsageResponse struct {
	UsageItems []UsageItem `json:"usageItems"`
}

// getBillingUsage fetches billing usage data from GitHub Enhanced Billing Platform API
func (c *BillingCollector) getBillingUsage() []UsageItem {
	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	var result []UsageItem

	// Fetch billing data for enterprises
	for _, name := range c.config.Enterprises {
		items := c.fetchBillingUsageForEntity(ctx, "enterprise", name)
		result = append(result, items...)
	}

	// Fetch billing data for organizations
	for _, name := range c.config.Orgs {
		items := c.fetchBillingUsageForEntity(ctx, "org", name)
		result = append(result, items...)
	}

	return result
}

// fetchBillingUsageForEntity fetches billing usage for a specific entity (org or enterprise)
func (c *BillingCollector) fetchBillingUsageForEntity(ctx context.Context, entityType, name string) []UsageItem {
	var endpoint string
	if entityType == "enterprise" {
		endpoint = fmt.Sprintf("/enterprises/%s/settings/billing/usage", name)
	} else {
		endpoint = fmt.Sprintf("/organizations/%s/settings/billing/usage", name)
	}

	req, err := c.client.NewRequest("GET", endpoint, nil)
	if err != nil {
		c.logger.Error("Failed to prepare billing request",
			"type", entityType,
			"name", name,
			"err", err,
		)
		c.failures.WithLabelValues("billing").Inc()
		return nil
	}

	response := &UsageResponse{}
	resp, err := c.client.Do(ctx, req, response)
	if err != nil {
		c.logger.Error("Failed to fetch billing usage",
			"type", entityType,
			"name", name,
			"err", err,
		)
		c.failures.WithLabelValues("billing").Inc()
		return nil
	}
	defer closeBody(resp)

	// Add metadata to each usage item
	var result []UsageItem
	for _, item := range response.UsageItems {
		item.Type = entityType
		item.Name = name
		result = append(result, item)
	}

	return result
}
