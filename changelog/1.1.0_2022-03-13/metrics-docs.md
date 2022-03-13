Change: Generate metrics documentation

We have added a script to automatically generate the available metrics within
the documentation to prevent any documentation gaps. Within the `hack/` folder
we got a small Go script which parses the available collectors and updates the
documentation partial based on that.

https://github.com/promhippie/github_exporter/issues/46
