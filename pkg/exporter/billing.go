package exporter

/*
GitHub Billing Collector v5.0.0 - Enhanced Billing Platform Support

This collector implements support for GitHub's Enhanced Billing Platform using
the unified billing API endpoints that replace the legacy product-specific endpoints.

API Endpoints:
- Organizations: GET /organizations/{org}/settings/billing/usage
- Enterprises:   GET /enterprises/{enterprise}/settings/billing/usage

Authentication Requirements:
- Organizations: Fine-grained OR Classic Personal Access Tokens supported
  - Fine-grained: "Administration" organization permissions (read)
  - Classic: Standard organization access
- Enterprises: Personal Access Tokens (Classic) ONLY
  - Fine-grained tokens NOT supported for enterprise endpoints
  - Classic PAT scope: "manage_billing:enterprise" required
- This creates different auth requirements between org and enterprise endpoints

Query Parameters Supported (configurable via config.Target.Billing):
- year: Filter by year (default: current year) - validated range 1900-2200
- month: Filter by month 1-12 (default: current month) - validated range 1-12
- day: Filter by day 1-31 (optional) - validated range 1-31
- hour: Filter by hour 0-23 (optional) - validated range 0-23
- cost_center_id: Filter by cost center (enterprises only, optional)

Breaking Changes from v4.x:
- Complete metric restructure with dimensional labels
- Repository-level attribution (cardinality limited to 100 per entity)
- Cost breakdown (gross, net, discount amounts)
- Date granularity for time-series analysis

Performance Considerations:
- Default query scope limited to current month to prevent large data downloads
- Repository cardinality limited to 100 repos per entity (excess aggregated as "other")
- Comprehensive data validation to prevent invalid metrics
*/

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

// BillingCollector collects metrics about the servers.
type BillingCollector struct {
	client   *github.Client
	logger   *slog.Logger
	db       store.Store
	failures *prometheus.CounterVec
	duration *prometheus.HistogramVec
	config   config.Target

	// v5.0.0 New granular billing metrics
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

	// v5.0.0 Enhanced label structure with configurable granularity
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
			"Usage quantity for GitHub products with repository-level attribution (v5.0.0+)",
			labels,
			nil,
		),
		UsageCostGross: prometheus.NewDesc(
			"github_billing_usage_cost_gross",
			"Gross cost before discounts for GitHub product usage (v5.0.0+)",
			labels,
			nil,
		),
		UsageCostNet: prometheus.NewDesc(
			"github_billing_usage_cost_net",
			"Net cost after discounts for GitHub product usage - actual charges (v5.0.0+)",
			labels,
			nil,
		),
		UsageCostDiscount: prometheus.NewDesc(
			"github_billing_usage_cost_discount",
			"Discount amount applied to GitHub product usage (v5.0.0+)",
			labels,
			nil,
		),
		UsagePricePerUnit: prometheus.NewDesc(
			"github_billing_usage_price_per_unit",
			"Price per unit for GitHub product usage (v5.0.0+)",
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

	c.logger.Debug("Fetched unified billing data",
		"count", len(usageData),
		"duration", time.Since(now),
	)

	for _, usage := range usageData {
		c.logger.Debug("Collecting billing usage",
			"type", usage.Type,
			"name", usage.Name,
			"product", usage.Product,
			"sku", usage.SKU,
			"date", usage.Date,
		)

		// Build labels based on granularity configuration
		labels := c.buildUsageLabels(usage)

		// Check if metric collection is enabled
		if !c.isMetricEnabled("quantity") {
			goto skipQuantity
		}
		ch <- prometheus.MustNewConstMetric(
			c.UsageQuantity,
			prometheus.GaugeValue,
			float64(usage.Quantity),
			labels...,
		)

	skipQuantity:
		if !c.isMetricEnabled("cost_gross") {
			goto skipGross
		}
		ch <- prometheus.MustNewConstMetric(
			c.UsageCostGross,
			prometheus.GaugeValue,
			usage.GrossAmount,
			labels...,
		)

	skipGross:
		if !c.isMetricEnabled("cost_net") {
			goto skipNet
		}
		ch <- prometheus.MustNewConstMetric(
			c.UsageCostNet,
			prometheus.GaugeValue,
			usage.NetAmount,
			labels...,
		)

	skipNet:
		if !c.isMetricEnabled("cost_discount") {
			goto skipDiscount
		}
		ch <- prometheus.MustNewConstMetric(
			c.UsageCostDiscount,
			prometheus.GaugeValue,
			usage.DiscountAmount,
			labels...,
		)

	skipDiscount:
		if !c.isMetricEnabled("price_per_unit") {
			goto skipPrice
		}
		ch <- prometheus.MustNewConstMetric(
			c.UsagePricePerUnit,
			prometheus.GaugeValue,
			usage.PricePerUnit,
			labels...,
		)

	skipPrice:
	}
}

// UnifiedUsageItem represents a single usage item with enhanced metadata for v5.0.0
type UnifiedUsageItem struct {
	Type             string  // "org" or "enterprise"
	Name             string  // Organization/enterprise name
	Date             string  // Usage date (YYYY-MM-DD)
	Product          string  // "Actions", "Packages", etc.
	SKU              string  // Product-specific SKU
	Quantity         int     // Usage quantity
	UnitType         string  // "minutes", "gigabytes", "GigabyteHours"
	PricePerUnit     float64 // Price per unit
	GrossAmount      float64 // Gross cost
	DiscountAmount   float64 // Discount applied
	NetAmount        float64 // Net cost (actual charge)
	OrganizationName string  // Organization name from API
	RepositoryName   string  // Repository name (format: "org/repo")
}

// getUnifiedBillingData fetches billing data from the unified API for all configured orgs and enterprises
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

// fetchBillingUsage fetches billing usage for a single organization or enterprise
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
		c.failures.WithLabelValues("billing_request_prep").Inc()
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

		// More granular error tracking with authentication hints
		if resp != nil {
			statusCode := resp.StatusCode
			if statusCode == 403 {
				if entityType == "enterprise" {
					c.logger.Error("Enterprise billing API access denied - ensure using Personal Access Token (Classic)",
						"type", entityType,
						"name", entityName,
						"status_code", statusCode,
						"note", "Fine-grained tokens NOT supported for enterprise billing endpoints",
						"required_scope", "manage_billing:enterprise",
					)
				} else {
					c.logger.Error("Organization billing API access denied - check token permissions",
						"type", entityType,
						"name", entityName,
						"status_code", statusCode,
						"note", "Fine-grained tokens supported with 'Administration' org permissions (read)",
						"classic_token_note", "Classic tokens also supported",
					)
				}
			}
			c.failures.WithLabelValues(fmt.Sprintf("billing_api_%d", statusCode)).Inc()
		} else {
			c.failures.WithLabelValues("billing_api_network").Inc()
		}
		return nil
	}
	defer closeBody(resp)

	c.logger.Debug("Fetched billing usage response",
		"type", entityType,
		"name", entityName,
		"usage_items", len(usage.UsageItems),
	)

	// Transform API response to internal format
	return c.transformUsageItems(usage.UsageItems, entityType, entityName)
}

// addBillingQueryParams adds query parameters based on configuration
func (c *BillingCollector) addBillingQueryParams(req *http.Request, entityType, entityName string) {
	q := req.URL.Query()
	now := time.Now()

	// Build query parameters based on configuration with validation
	params := make(map[string]string)

	// Year parameter (default to current year if not specified)
	if c.config.Billing.Year != nil {
		year := *c.config.Billing.Year
		if year < 1900 || year > 2200 { // Reasonable bounds
			c.logger.Warn("Invalid year parameter, using current year",
				"configured_year", year,
				"current_year", now.Year(),
			)
			params["year"] = fmt.Sprintf("%d", now.Year())
		} else {
			params["year"] = fmt.Sprintf("%d", year)
		}
	} else {
		// Default to current year
		params["year"] = fmt.Sprintf("%d", now.Year())
	}

	// Month parameter (default to current month if not specified for performance)
	if c.config.Billing.Month != nil {
		month := *c.config.Billing.Month
		if month < 1 || month > 12 {
			c.logger.Warn("Invalid month parameter, using current month",
				"configured_month", month,
				"current_month", int(now.Month()),
			)
			params["month"] = fmt.Sprintf("%d", int(now.Month()))
		} else {
			params["month"] = fmt.Sprintf("%d", month)
		}
	} else {
		// Default to current month to prevent massive data downloads
		params["month"] = fmt.Sprintf("%d", int(now.Month()))
	}

	// Day parameter (optional, validated range 1-31)
	if c.config.Billing.Day != nil {
		day := *c.config.Billing.Day
		if day < 1 || day > 31 {
			c.logger.Warn("Invalid day parameter, skipping day filter",
				"configured_day", day,
			)
		} else {
			params["day"] = fmt.Sprintf("%d", day)
		}
	}

	// Hour parameter (optional, validated range 0-23)
	if c.config.Billing.Hour != nil {
		hour := *c.config.Billing.Hour
		if hour < 0 || hour > 23 {
			c.logger.Warn("Invalid hour parameter, skipping hour filter",
				"configured_hour", hour,
			)
		} else {
			params["hour"] = fmt.Sprintf("%d", hour)
		}
	}

	// Cost center ID (enterprises only)
	if c.config.Billing.CostCenterID != nil && entityType == "enterprise" {
		costCenterID := *c.config.Billing.CostCenterID
		if costCenterID != "" { // Only add if not empty
			params["cost_center_id"] = costCenterID
		}
	} else if c.config.Billing.CostCenterID != nil && entityType == "org" {
		c.logger.Debug("Cost center ID specified but not supported for organization endpoints",
			"type", entityType,
			"name", entityName,
		)
	}

	// Add all parameters to the request
	for key, value := range params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	// Log the applied parameters
	logFields := []any{
		"type", entityType,
		"name", entityName,
	}
	for key, value := range params {
		logFields = append(logFields, key, value)
	}

	c.logger.Debug("Added billing query parameters", logFields...)
}

// transformUsageItems converts API usage items to internal UnifiedUsageItem format
func (c *BillingCollector) transformUsageItems(items []UsageItem, entityType, entityName string) []*UnifiedUsageItem {
	result := make([]*UnifiedUsageItem, 0, len(items))
	repoCount := make(map[string]bool)

	// Use configurable repository limit or default
	maxRepositories := 100 // Default
	if c.config.Billing.MaxRepositories != nil {
		maxRepositories = *c.config.Billing.MaxRepositories
	}

	for _, item := range items {
		// Validate basic data integrity
		if !c.isValidUsageItem(item) {
			c.logger.Warn("Skipping invalid usage item",
				"type", entityType,
				"name", entityName,
				"product", item.Product,
				"sku", item.SKU,
				"quantity", item.Quantity,
			)
			continue
		}

		// Handle repository cardinality limiting
		repositoryName := c.normalizeRepositoryName(item.RepositoryName, repoCount, maxRepositories)

		// Ensure organization name is set
		organizationName := item.OrganizationName
		if organizationName == "" {
			organizationName = entityName
		}

		unifiedItem := &UnifiedUsageItem{
			Type:             entityType,
			Name:             entityName,
			Date:             item.Date,
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

	if len(repoCount) > maxRepositories {
		c.logger.Warn("Repository cardinality limited",
			"type", entityType,
			"name", entityName,
			"total_repos", len(repoCount),
			"limit", maxRepositories,
		)
	}

	return result
}

// isValidUsageItem validates usage item data integrity
func (c *BillingCollector) isValidUsageItem(item UsageItem) bool {
	// Check for required fields
	if item.Product == "" || item.SKU == "" || item.UnitType == "" {
		return false
	}

	// Check for valid quantities
	if item.Quantity < 0 {
		return false
	}

	// Check for valid costs (can be 0, but not negative)
	if item.GrossAmount < 0 || item.NetAmount < 0 || item.DiscountAmount < 0 {
		return false
	}

	// Check for valid price per unit
	if item.PricePerUnit < 0 {
		return false
	}

	// Basic date format validation (YYYY-MM-DD expected)
	if len(item.Date) != 10 || item.Date[4] != '-' || item.Date[7] != '-' {
		return false
	}

	return true
}

// normalizeRepositoryName handles repository cardinality limiting and normalization
func (c *BillingCollector) normalizeRepositoryName(repoName string, repoCount map[string]bool, maxRepos int) string {
	// Handle empty repository names
	if repoName == "" {
		return "unknown"
	}

	// If we're under the limit, use the actual repo name
	if len(repoCount) < maxRepos {
		repoCount[repoName] = true
		return repoName
	}

	// If we've seen this repo before, continue using it
	if _, exists := repoCount[repoName]; exists {
		return repoName
	}

	// If we're over the limit and this is a new repo, aggregate it
	return "other"
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

// buildMetricLabels creates the label slice based on granularity configuration
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

// buildUsageLabels creates metric labels for a specific usage item based on configuration
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

// isMetricEnabled checks if a specific metric type should be collected
func (c *BillingCollector) isMetricEnabled(metricType string) bool {
	// If no specific metrics are configured, all are enabled by default
	if len(c.config.Billing.EnabledMetrics) == 0 {
		return true
	}

	// Check if this metric type is in the enabled list
	for _, enabled := range c.config.Billing.EnabledMetrics {
		if enabled == metricType {
			return true
		}
	}

	return false
}
