Bugfix: Prevent concurrent scrapes

If the exporter got some kind of duplicated repository names configured it lead
to errors because the label combination had been scraped already. We have added
some simple checks to prevent duplicated exports an all currently available
collectors.

https://github.com/promhippie/github_exporter/issues/190
