Enhancement: Use getter functions to get values

To reduce the used boilerplate code and to better use the GitHub library we have
updated most of the available collectors to simply use the provided getter
functions instead of checking for nil values everywhere on our own.

https://github.com/promhippie/github_exporter/pull/183
