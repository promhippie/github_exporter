Bugfix: Fixed typecasts within billing API

In some cases it happened that the GitHub billing API responded with floats, but
we only accepted integers. With this fix any number should be casted to floats
all the time.

https://github.com/promhippie/github_exporter/issues/67
