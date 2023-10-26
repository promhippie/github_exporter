Change: Read secrets form files

We have added proper support to load secrets like the token or the private key
for app authentication from files or from base64-encoded strings. Just provide
the flags or environment variables for token or private key with a DSN formatted
string like `file://path/to/file` or `base64://Zm9vYmFy`.

Since the private key for GitHub App authentication had been provided in
base64-encoded format this is a breaking change as this won't work anymore until
you prefix the value with `base64://`.

https://github.com/promhippie/github_exporter/pull/245
