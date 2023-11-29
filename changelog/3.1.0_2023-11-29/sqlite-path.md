Bugfix: Create SQLite directory if it doesn't exist

We have integrated a fix to always try to create the directory for the SQLite
database if it doesn't exist. The exporter will fail to start if the directory
does not exist and if it fails to create the directory for the database file.

https://github.com/promhippie/github_exporter/issues/278
