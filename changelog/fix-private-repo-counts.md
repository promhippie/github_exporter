Fix: Handle private repos counts not available to non-organization member
[#18](https://github.com/promhippie/github_exporter/issues/18)

Bug fix the case where the account used to query Github API is not a member of a given
organisation. The resulted in segfault before the fix
