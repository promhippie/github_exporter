Bugfix: Resolve nil pointer issues with responses

We have introduced previously a feature where we made sure to close all response
bodies, but we forgot to properly check if these responses are really valid
which lead to nil pointer dereference issues.

https://github.com/promhippie/github_exporter/issues/216
