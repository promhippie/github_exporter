Bugfix: Handle private repos counts not available to non-organization member

Fix the case where the account used to query GitHub API is not a member of a
given organisation, it resulted in a segfault.

https://github.com/promhippie/github_exporter/pull/18
