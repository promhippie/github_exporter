Bugfix: Fixed pointer references within exporters

So far we directly accessed common attributes on repos within the collectors,
but as there could be corner cases where some attribute could be missing we add
conditions to make sure only set values are getting called.

https://github.com/promhippie/github_exporter/issues/49
