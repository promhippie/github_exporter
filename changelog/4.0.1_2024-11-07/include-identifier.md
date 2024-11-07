Bugfix: Include identifier in labels for workflow jobs

To avoid errors related to already scraped metrics we have added the identifier
to the default labels for the workflow job collector.

https://github.com/promhippie/github_exporter/pull/418
