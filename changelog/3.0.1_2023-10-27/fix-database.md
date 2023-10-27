Bugfix: Correctly store and retrieve records

We had introduced a bug while switching between golang's sqlx and sql packages,
with this fix all workflows should be stored and retrieved correctly.

https://github.com/promhippie/github_exporter/issues/270
