Bugfix: Fix PostgreSQL workflow_job identifiers to be bigint

github ids are 64 bit integers, so pg can't fit them in an int

https://github.com/promhippie/github_exporter/pull/414
