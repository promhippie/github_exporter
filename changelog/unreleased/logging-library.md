Change: Switch to official logging library

Since there have been a structured logger part of the Go standard library we
thought it's time to replace the library with that. Be aware that log messages
should change a little bit.

https://github.com/promhippie/github_exporter/issues/393
