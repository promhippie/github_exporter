Bugfix: Update conclusion and completed_at columns

For the new workflow job collector we had been missing the `conclusion` and
`completed_at` values, they will be stored by future webhook events.

https://github.com/promhippie/github_exporter/pull/418
