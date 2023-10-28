Bugfix: Correctly store and retrieve records

We had introduced a bug while switching between golang's sqlx and sql packages,
with this fix all workflows should be stored and retrieved correctly.

You got to make sure to delete the database which have been created with the
3.0.0 release as the migration setup have been changed.

https://github.com/promhippie/github_exporter/issues/270
