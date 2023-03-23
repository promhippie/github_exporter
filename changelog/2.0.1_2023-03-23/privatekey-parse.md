Bugfix: Improve parsing of private key for GitHub app

Previously we always checked if a file with the value exists if a private key
for GitHub app authentication have been provided, now I have switched it to try
to base64 decode the string first, and try to load the file afterwards which
works more reliable and avoids leaking the private key into the log output.

https://github.com/promhippie/github_exporter/issues/185
