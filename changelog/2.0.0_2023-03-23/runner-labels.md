Bugfix: Set right name/owner labels for runner metrics

We introduced metrics for GitHub self-hosted runners but we missed some
important labels as remaining todos. With this change this gets corrected to
properly show the repo/org/enterprise where the runner have been attached to.

https://github.com/promhippie/github_exporter/pull/184
