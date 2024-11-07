Bugfix: Fix PostgreSQL identifiers to be bigint

Since Github submit 64bit integers for the identifier of workflow jobs we had to
fix the type to `BIGINT` for the database schema to avoid errors related to
store events.

https://github.com/promhippie/github_exporter/pull/414
