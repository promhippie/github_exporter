Bugfix: Missing field updates for workflow jobs

With the previous behavior the runner specific labels for worklow job metrics
have always been empty. They we always set after the job started but have never
been part of the database table update process.

https://github.com/promhippie/github_exporter/issues/485
