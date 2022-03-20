Bugfix: Fixed wildcard matching for private repos

We have fixed the ability to match by globbing/wildcard for private
repositories. So far the private repositories have been missing with the latest
globbing/wildcard matching.

https://github.com/promhippie/github_exporter/issues/75
