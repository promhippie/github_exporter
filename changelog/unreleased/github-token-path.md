Enhancement: Allow reading github token from path

The `--github.token` now optionally accepts a file path as an argument.
The program will read the contents of the file, and use it as the token.

https://github.com/promhippie/github_exporter/issues/238