# GitHub Exporter

[![Current Tag](https://img.shields.io/github/v/tag/promhippie/github_exporter?sort=semver)](https://github.com/promhippie/github_exporter) [![General Build](https://github.com/promhippie/github_exporter/actions/workflows/general.yml/badge.svg)](https://github.com/promhippie/github_exporter/actions/workflows/general.yml) [![Join the Matrix chat at https://matrix.to/#/#webhippie:matrix.org](https://img.shields.io/badge/matrix-%23webhippie-7bc9a4.svg)](https://matrix.to/#/#webhippie:matrix.org) [![Codacy Badge](https://app.codacy.com/project/badge/Grade/af9b80ac46294ac9a52d823e991eb4e9)](https://app.codacy.com/gh/promhippie/github_exporter/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade) [![Go Doc](https://godoc.org/github.com/promhippie/github_exporter?status.svg)](http://godoc.org/github.com/promhippie/github_exporter) [![Go Report](http://goreportcard.com/badge/github.com/promhippie/github_exporter)](http://goreportcard.com/report/github.com/promhippie/github_exporter)

An exporter for [Prometheus][prometheus] that collects metrics from
[GitHub][github].

## ⚠️ Breaking Changes in v5.0.0

**Enhanced Billing Platform Support**: v5.0.0 introduces major changes to billing metrics collection due to GitHub's migration to the Enhanced Billing Platform.

### Key Changes:
- **New Unified Metrics**: Replaced 10+ legacy billing metrics with 5 comprehensive dimensional metrics
- **Enhanced API Support**: Updated to use `/settings/billing/usage` endpoints (legacy endpoints return 410 errors)
- **Repository Attribution**: Granular usage tracking per repository with cardinality controls  
- **Cost Breakdown**: Separate metrics for gross, net, and discount amounts
- **Query Parameters**: Configurable time ranges (year/month/day/hour) and cost center filtering

### New v5.0.0 Billing Metrics:
```
github_billing_usage_quantity          # Usage quantities with full dimensional breakdown
github_billing_usage_cost_gross        # Gross costs before discounts  
github_billing_usage_cost_net          # Net costs after discounts (actual charges)
github_billing_usage_cost_discount     # Discount amounts applied
github_billing_usage_price_per_unit    # Per-unit pricing information
```

### Migration Required:
- **Dashboards**: Update Grafana/visualization queries for new metric structure
- **Alerts**: Convert alerting rules to use new dimensional metrics
- **Authentication**: Enterprise billing requires Personal Access Tokens (Classic) only

See CHANGELOG.md for migration requirements and breaking changes details.

## Install

You can download prebuilt binaries from our [GitHub releases][releases], or you
can use our containers published on [Docker Hub][dockerhub] and [Quay][quayio].
If you need further guidance how to install this take a look at our
[documentation][docs].

## Development

Make sure you have a working Go environment, for further reference or a guide
take a look at the [install instructions][golang]. This project requires
Go >= v1.17, at least that's the version we are using.

```console
git clone https://github.com/promhippie/github_exporter.git
cd github_exporter

make generate build

./bin/github_exporter -h
```

## Enhanced Billing Configuration (v5.0.0+)

Configure granular billing data collection with new options:

```yaml
# Basic billing collection
target:
  enterprises:
    - "my-enterprise"
  orgs:
    - "my-org"

# Advanced configuration with query parameters and granularity controls
target:
  billing:
    # Query parameter filtering
    year: 2024                           # Filter by specific year  
    month: 12                            # Filter by specific month
    cost_center_id: "engineering"        # Enterprise cost center filtering
    
    # Performance and cardinality controls  
    max_repositories: 50                 # Limit repository cardinality
    disable_repository_labels: false     # Disable repo-level attribution
    disable_date_labels: false           # Disable temporal labels
    enabled_metrics:                     # Select specific metrics
      - "quantity"
      - "cost_net"
  enterprises:
    - "my-enterprise"
```

**Authentication Notes**:
- **Organizations**: Support both Fine-grained and Classic Personal Access Tokens
- **Enterprises**: Require Classic Personal Access Tokens only (`manage_billing:enterprise` scope)

See the configuration examples below and auto-generated metrics documentation for details.

### Example Prometheus Queries

```promql
# Total monthly GitHub Actions cost across all organizations
sum(github_billing_usage_cost_net{product="Actions"}) by (organization)

# Actions usage by repository (top 10)
topk(10, sum(github_billing_usage_quantity{product="Actions"}) by (repository))

# Total savings from discounts
sum(github_billing_usage_cost_discount) by (type, name)

# Cost efficiency: usage per dollar spent
sum(github_billing_usage_quantity{product="Actions"}) / sum(github_billing_usage_cost_net{product="Actions"})

# Daily cost trend for specific organization  
sum(github_billing_usage_cost_net{organization="my-org"}) by (date)
```

## Security

If you find a security issue please contact
[thomas@webhippie.de](mailto:thomas@webhippie.de) first.

## Contributing

Fork -> Patch -> Push -> Pull Request

## Authors

-   [Thomas Boerger](https://github.com/tboerger)

## License

Apache-2.0

## Copyright

```console
Copyright (c) 2018 Thomas Boerger <thomas@webhippie.de>
```

[prometheus]: https://prometheus.io
[github]: https://github.com
[releases]: https://github.com/promhippie/github_exporter/releases
[dockerhub]: https://hub.docker.com/r/promhippie/github-exporter/tags/
[quayio]: https://quay.io/repository/promhippie/github-exporter?tab=tags
[docs]: https://promhippie.github.io/github_exporter/#getting-started
[golang]: http://golang.org/doc/install.html
