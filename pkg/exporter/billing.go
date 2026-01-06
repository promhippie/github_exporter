package exporter

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/go-github/v81/github"
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

	Usage          *prometheus.Desc
	GrossAmount    *prometheus.Desc
	DiscountAmount *prometheus.Desc
	NetAmount      *prometheus.Desc
	PricePerUnit   *prometheus.Desc
}

// NewBillingCollector returns a new BillingCollector.
func NewBillingCollector(logger *slog.Logger, client *github.Client, db store.Store, failures *prometheus.CounterVec, duration *prometheus.HistogramVec, cfg config.Target) *BillingCollector {
	if failures != nil {
		failures.WithLabelValues("billing").Add(0)
	}

	labels := []string{"type", "name", "product", "sku", "unit", "date", "org", "repo"}
	return &BillingCollector{
		client:   client,
		logger:   logger.With("collector", "billing"),
		db:       db,
		failures: failures,
		duration: duration,
		config:   cfg,

		Usage: prometheus.NewDesc(
			"github_billing_usage",
			"Usage quantity from GitHub Enhanced Billing Platform",
			labels,
			nil,
		),
		GrossAmount: prometheus.NewDesc(
			"github_billing_usage_gross_amount",
			"Gross amount charged for this usage item",
			labels,
			nil,
		),
		DiscountAmount: prometheus.NewDesc(
			"github_billing_usage_discount_amount",
			"Discount amount applied to this usage item",
			labels,
			nil,
		),
		NetAmount: prometheus.NewDesc(
			"github_billing_usage_net_amount",
			"Net amount charged for this usage item after discounts",
			labels,
			nil,
		),
		PricePerUnit: prometheus.NewDesc(
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
		c.Usage,
		c.GrossAmount,
		c.DiscountAmount,
		c.NetAmount,
		c.PricePerUnit,
	}
}

// Describe sends the super-set of all possible descriptors of metrics collected by this Collector.
func (c *BillingCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Usage
	ch <- c.GrossAmount
	ch <- c.DiscountAmount
	ch <- c.NetAmount
	ch <- c.PricePerUnit
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *BillingCollector) Collect(ch chan<- prometheus.Metric) {
	collected := make(map[string]bool)
	now := time.Now()
	usage := c.getUsage()
	c.duration.WithLabelValues("billing").Observe(time.Since(now).Seconds())

	c.logger.Debug("Fetched billing usage",
		"count", len(usage),
		"duration", time.Since(now),
	)

	for _, item := range usage {
		key := fmt.Sprintf(
			"%s-%s-%s-%s-%s",
			item.Type,
			item.Name,
			item.Product,
			item.SKU,
			item.Date,
		)

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
			"unit", item.UnitType,
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
			c.Usage,
			prometheus.GaugeValue,
			item.Quantity,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GrossAmount,
			prometheus.GaugeValue,
			item.GrossAmount,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DiscountAmount,
			prometheus.GaugeValue,
			item.DiscountAmount,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NetAmount,
			prometheus.GaugeValue,
			item.NetAmount,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PricePerUnit,
			prometheus.GaugeValue,
			item.PricePerUnit,
			labels...,
		)
	}
}

// UsageItem represents a billing usage item from GitHub Enhanced Billing Platform API.
type UsageItem struct {
	Type             string
	Name             string
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
}

// UsageResponse represents the response from GitHub Enhanced Billing Platform API.
type UsageResponse struct {
	UsageItems []UsageItem `json:"usageItems"`
}

// getUsage fetches billing usage data from GitHub Enhanced Billing Platform API.
func (c *BillingCollector) getUsage() []UsageItem {
	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	var result []UsageItem

	for _, name := range c.config.Enterprises {
		items := c.fetchUsageForEntity(ctx, "enterprise", name)
		result = append(result, items...)
	}

	for _, name := range c.config.Orgs {
		items := c.fetchUsageForEntity(ctx, "org", name)
		result = append(result, items...)
	}

	return result
}

// fetchUsageForEntity fetches billing usage for a specific entity (org or enterprise).
func (c *BillingCollector) fetchUsageForEntity(ctx context.Context, entity, name string) []UsageItem {
	var (
		endpoint string
		result   []UsageItem
	)

	if entity == "enterprise" {
		endpoint = fmt.Sprintf("/enterprises/%s/settings/billing/usage", name)
	} else {
		endpoint = fmt.Sprintf("/organizations/%s/settings/billing/usage", name)
	}

	req, err := c.client.NewRequest("GET", endpoint, nil)

	if err != nil {
		c.logger.Error("Failed to prepare billing usage",
			"type", entity,
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
			"type", entity,
			"name", name,
			"err", err,
		)

		c.failures.WithLabelValues("billing").Inc()
		return nil
	}

	defer closeBody(resp)

	for _, item := range response.UsageItems {
		item.Type = entity
		item.Name = name

		result = append(result, item)
	}

	return result
}
