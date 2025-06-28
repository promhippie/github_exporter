Change: Change billing metrics structure for matching the new API

**BREAKING CHANGE**: This is a major version bump (v5.0.0) with complete redesign of billing metrics.

Old v4.x Metrics (REMOVED):
```
github_action_billing_minutes_used
github_package_billing_gigabytes_bandwidth_used  
github_storage_billing_estimated_storage_for_month
```

New v5.0.0 Metrics:
```
github_billing_usage                    # Usage quantity
github_billing_usage_gross_amount       # Gross amount charged
github_billing_usage_discount_amount    # Discount amount applied  
github_billing_usage_net_amount         # Net amount after discounts
github_billing_usage_price_per_unit     # Price per unit
```

The new metrics provide granular repository-level tracking with rich labels including:
- `date`: Timestamp of usage
- `product`: Service type (actions, packages, git_lfs)
- `sku`: Detailed SKU information
- `unit_type`: Unit of measurement
- `repository_name`: Specific repository
- `organization_name`: Organization name

https://github.com/promhippie/github_exporter/issues/496