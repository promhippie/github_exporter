package exporter

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/go-github/v72/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/promhippie/github_exporter/pkg/config"
	"github.com/promhippie/github_exporter/pkg/store"
)

// BillingCollector collects billing metrics from the unified billing API.
type BillingCollector struct {
	client   *github.Client
	logger   *slog.Logger
	db       store.Store
	failures *prometheus.CounterVec
	duration *prometheus.HistogramVec
	config   config.Target

	UsageQuantity     *prometheus.Desc
	UsageCostGross    *prometheus.Desc
	UsageCostNet      *prometheus.Desc
	UsageCostDiscount *prometheus.Desc
	UsagePricePerUnit *prometheus.Desc
}

// NewBillingCollector returns a new BillingCollector.
func NewBillingCollector(logger *slog.Logger, client *github.Client, db store.Store, failures *prometheus.CounterVec, duration *prometheus.HistogramVec, cfg config.Target) *BillingCollector {
	if failures != nil {
		failures.WithLabelValues("billing").Add(0)
	}

	labels := buildMetricLabels(cfg.Billing)

	return &BillingCollector{
		client:   client,
		logger:   logger.With("collector", "billing"),
		db:       db,
		failures: failures,
		duration: duration,
		config:   cfg,

		UsageQuantity: prometheus.NewDesc(
			"github_billing_usage_quantity",
			"Usage quantity for GitHub products",
			labels,
			nil,
		),
		UsageCostGross: prometheus.NewDesc(
			"github_billing_usage_cost_gross",
			"Gross cost before discounts for GitHub product usage",
			labels,
			nil,
		),
		UsageCostNet: prometheus.NewDesc(
			"github_billing_usage_cost_net",
			"Net cost after discounts for GitHub product usage",
			labels,
			nil,
		),
		UsageCostDiscount: prometheus.NewDesc(
			"github_billing_usage_cost_discount",
			"Discount amount applied to GitHub product usage",
			labels,
			nil,
		),
		UsagePricePerUnit: prometheus.NewDesc(
			"github_billing_usage_price_per_unit",
			"Price per unit for GitHub product usage",
			labels,
			nil,
		),
	}
}

// Metrics simply returns the list metric descriptors for generating a documentation.
func (c *BillingCollector) Metrics() []*prometheus.Desc {
	return []*prometheus.Desc{
		c.UsageQuantity,
		c.UsageCostGross,
		c.UsageCostNet,
		c.UsageCostDiscount,
		c.UsagePricePerUnit,
	}
}

// Describe sends the super-set of all possible descriptors of metrics collected by this Collector.
func (c *BillingCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.UsageQuantity
	ch <- c.UsageCostGross
	ch <- c.UsageCostNet
	ch <- c.UsageCostDiscount
	ch <- c.UsagePricePerUnit
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *BillingCollector) Collect(ch chan<- prometheus.Metric) {
	now := time.Now()
	usageData := c.getUnifiedBillingData()
	c.duration.WithLabelValues("billing").Observe(time.Since(now).Seconds())

	c.logger.Debug("Fetched billing data",
		"count", len(usageData),
		"duration", time.Since(now),
	)

	for _, usage := range usageData {
		labels := c.buildUsageLabels(usage)

		if c.isMetricEnabled("quantity") {
			ch <- prometheus.MustNewConstMetric(
				c.UsageQuantity,
				prometheus.GaugeValue,
				float64(usage.Quantity),
				labels...,
			)
		}

		if c.isMetricEnabled("cost_gross") {
			ch <- prometheus.MustNewConstMetric(
				c.UsageCostGross,
				prometheus.GaugeValue,
				usage.GrossAmount,
				labels...,
			)
		}

		if c.isMetricEnabled("cost_net") {
			ch <- prometheus.MustNewConstMetric(
				c.UsageCostNet,
				prometheus.GaugeValue,
				usage.NetAmount,
				labels...,
			)
		}

		if c.isMetricEnabled("cost_discount") {
			ch <- prometheus.MustNewConstMetric(
				c.UsageCostDiscount,
				prometheus.GaugeValue,
				usage.DiscountAmount,
				labels...,
			)
		}

		if c.isMetricEnabled("price_per_unit") {
			ch <- prometheus.MustNewConstMetric(
				c.UsagePricePerUnit,
				prometheus.GaugeValue,
				usage.PricePerUnit,
				labels...,
			)
		}
	}
}

// UnifiedUsageItem represents a single usage item.
type UnifiedUsageItem struct {
	Type             string
	Name             string
	Date             string
	Product          string
	SKU              string
	Quantity         float64
	UnitType         string
	PricePerUnit     float64
	GrossAmount      float64
	DiscountAmount   float64
	NetAmount        float64
	OrganizationName string
	RepositoryName   string
}

func (c *BillingCollector) getUnifiedBillingData() []*UnifiedUsageItem {
	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	result := make([]*UnifiedUsageItem, 0)

	// Fetch enterprise billing data
	for _, name := range c.config.Enterprises {
		items := c.fetchBillingUsage(ctx, "enterprise", name)
		result = append(result, items...)
	}

	// Fetch organization billing data
	for _, name := range c.config.Orgs {
		items := c.fetchBillingUsage(ctx, "org", name)
		result = append(result, items...)
	}

	return result
}

func (c *BillingCollector) fetchBillingUsage(ctx context.Context, entityType, entityName string) []*UnifiedUsageItem {
	var endpoint string
	if entityType == "enterprise" {
		endpoint = fmt.Sprintf("/enterprises/%s/settings/billing/usage", entityName)
	} else {
		endpoint = fmt.Sprintf("/organizations/%s/settings/billing/usage", entityName)
	}

	req, err := c.client.NewRequest("GET", endpoint, nil)
	if err != nil {
		c.logger.Error("Failed to prepare billing request",
			"type", entityType,
			"name", entityName,
			"err", err,
		)
		c.failures.WithLabelValues("billing").Inc()
		return nil
	}

	// Add query parameters based on configuration
	c.addBillingQueryParams(req, entityType, entityName)

	usage := &UnifiedBillingResponse{}
	resp, err := c.client.Do(ctx, req, usage)
	if err != nil {
		c.logger.Error("Failed to fetch billing usage",
			"type", entityType,
			"name", entityName,
			"err", err,
		)

		c.failures.WithLabelValues("billing").Inc()
		return nil
	}
	defer closeBody(resp)

	return c.transformUsageItems(usage.UsageItems, entityType, entityName)
}

func (c *BillingCollector) addBillingQueryParams(req *http.Request, entityType, _ string) {
	q := req.URL.Query()
	now := time.Now()

	params := make(map[string]string)
	if c.config.Billing.Year != nil {
		year := *c.config.Billing.Year
		if year < 1900 || year > 2200 {
			params["year"] = fmt.Sprintf("%d", now.Year())
		} else {
			params["year"] = fmt.Sprintf("%d", year)
		}
	} else {
		params["year"] = fmt.Sprintf("%d", now.Year())
	}

	if c.config.Billing.Month != nil {
		month := *c.config.Billing.Month
		if month < 1 || month > 12 {
			params["month"] = fmt.Sprintf("%d", int(now.Month()))
		} else {
			params["month"] = fmt.Sprintf("%d", month)
		}
	} else {
		params["month"] = fmt.Sprintf("%d", int(now.Month()))
	}

	if c.config.Billing.Day != nil {
		day := *c.config.Billing.Day
		if day >= 1 && day <= 31 {
			params["day"] = fmt.Sprintf("%d", day)
		}
	}

	if c.config.Billing.Hour != nil {
		hour := *c.config.Billing.Hour
		if hour >= 0 && hour <= 23 {
			params["hour"] = fmt.Sprintf("%d", hour)
		}
	}

	if c.config.Billing.CostCenterID != nil && entityType == "enterprise" {
		costCenterID := *c.config.Billing.CostCenterID
		if costCenterID != "" {
			params["cost_center_id"] = costCenterID
		}
	}

	for key, value := range params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()
}

func (c *BillingCollector) transformUsageItems(items []UsageItem, entityType, entityName string) []*UnifiedUsageItem {
	result := make([]*UnifiedUsageItem, 0, len(items))
	repoCount := make(map[string]bool)

	maxRepositories := 100
	if c.config.Billing.MaxRepositories != nil {
		maxRepositories = *c.config.Billing.MaxRepositories
	}

	for _, item := range items {
		if !c.isValidUsageItem(item) {
			continue
		}

		repositoryName := c.normalizeRepositoryName(item.RepositoryName, repoCount, maxRepositories)

		organizationName := item.OrganizationName
		if organizationName == "" {
			organizationName = entityName
		}

		unifiedItem := &UnifiedUsageItem{
			Type:             entityType,
			Name:             entityName,
			Date:             c.normalizeDate(item.Date),
			Product:          item.Product,
			SKU:              item.SKU,
			Quantity:         item.Quantity,
			UnitType:         item.UnitType,
			PricePerUnit:     item.PricePerUnit,
			GrossAmount:      item.GrossAmount,
			DiscountAmount:   item.DiscountAmount,
			NetAmount:        item.NetAmount,
			OrganizationName: organizationName,
			RepositoryName:   repositoryName,
		}

		result = append(result, unifiedItem)
	}

	return result
}

func (c *BillingCollector) isValidUsageItem(item UsageItem) bool {
	if item.Product == "" || item.SKU == "" || item.UnitType == "" {
		return false
	}

	if item.Quantity < 0 || item.GrossAmount < 0 || item.NetAmount < 0 || item.DiscountAmount < 0 || item.PricePerUnit < 0 {
		return false
	}

	if item.Date == "" || len(item.Date) < 10 || item.Date[4] != '-' || item.Date[7] != '-' {
		return false
	}

	return true
}

func (c *BillingCollector) normalizeRepositoryName(repoName string, repoCount map[string]bool, maxRepos int) string {
	if repoName == "" {
		return "unknown"
	}

	if len(repoCount) < maxRepos {
		repoCount[repoName] = true
		return repoName
	}

	if _, exists := repoCount[repoName]; exists {
		return repoName
	}

	return "other"
}

// UnifiedBillingResponse represents the unified billing API response.
type UnifiedBillingResponse struct {
	UsageItems []UsageItem `json:"usageItems"`
}

// UsageItem represents an individual usage item.
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
}

func buildMetricLabels(billingConfig config.Billing) []string {
	labels := []string{"type", "name", "product", "sku", "unit_type"}

	if !billingConfig.DisableDateLabels {
		labels = append(labels, "date")
	}

	if !billingConfig.DisableOrganizationLabels {
		labels = append(labels, "organization")
	}

	if !billingConfig.DisableRepositoryLabels {
		labels = append(labels, "repository")
	}

	return labels
}

func (c *BillingCollector) buildUsageLabels(usage *UnifiedUsageItem) []string {
	labels := []string{
		usage.Type,     // type
		usage.Name,     // name
		usage.Product,  // product
		usage.SKU,      // sku
		usage.UnitType, // unit_type
	}

	if !c.config.Billing.DisableDateLabels {
		labels = append(labels, usage.Date)
	}

	if !c.config.Billing.DisableOrganizationLabels {
		labels = append(labels, usage.OrganizationName)
	}

	if !c.config.Billing.DisableRepositoryLabels {
		labels = append(labels, usage.RepositoryName)
	}

	return labels
}

func (c *BillingCollector) isMetricEnabled(metricType string) bool {
	if len(c.config.Billing.EnabledMetrics) == 0 {
		return true
	}

	for _, enabled := range c.config.Billing.EnabledMetrics {
		if enabled == metricType {
			return true
		}
	}

	return false
}

func (c *BillingCollector) normalizeDate(dateStr string) string {
	if len(dateStr) == 10 {
		return dateStr
	}

	if len(dateStr) >= 10 && dateStr[4] == '-' && dateStr[7] == '-' {
		return dateStr[:10]
	}

	return dateStr
}
