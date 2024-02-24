Bugfix: Drop out of order event if same timestamp as existing completed event

We've fixed the behavior of out of order workflow events with the same
timestamp, by preferring an already recorded `completed` event

https://github.com/promhippie/github_exporter/issues/296
https://github.com/promhippie/github_exporter/pull/298
https://github.com/promhippie/github_exporter/pull/312

