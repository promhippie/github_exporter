Enhancement: Integrate option pprof profiling

We have added an option to enable a pprof endpoint for proper profiling support
with the help of tools like Parca. The endpoint `/debug/pprof` can now
optionally be enabled to get the profiling details for catching potential memory
leaks.

https://github.com/promhippie/github_exporter/pull/193
