Enhancement: Rebuild workflow collector based on webhooks

We have rebuilt the workflow collector based on GitHub webhooks to get a more
stable behavior and to avoid running into many rate limits. Receiving changes
directly from gitHub instead of requesting it should improve the behavior a lot.
Please read the docs to see how you can setup the required webhooks within
GitHub, opening up the exporter for GitHub should be the most simple part.

https://github.com/promhippie/github_exporter/issues/261
