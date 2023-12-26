Bugfix: Drop out of order workflow webhook events

We've fixed the behavior of out of order workflow events by dropping them,
since the events order is not guaranteed by GitHub

https://github.com/promhippie/github_exporter/issues/296
https://github.com/promhippie/github_exporter/pull/298
