Bugfix: Add SQLite and Genji if supported

We haven't been able to build all supported binaries as some have been lacking
support for the use libraries for SQLite and Genji. We have added build tags
which enables or disables the database drivers if needed.

https://github.com/promhippie/github_exporter/pull/272
