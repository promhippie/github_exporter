---
title: "Webhook"
date: 2023-10-20T00:00:00+00:00
anchor: "webhook"
weight: 40
---

For the configuration of a webhook on your GitHub repository or your
organization you should just follow the defined steps, after that you should
receive any webhook related to the defined actions on the exporter if you have
followed the instructions from above.

If you want to configure a webhook for your organization just visit
`https://github.com/organizations/ORGANIZATION/settings/hooks/new` where you got
to replace `ORGANIZATION` with the name of your organization, after that follow
the steps from the screenshots.

If you want to configure a webhook for your repository just visit
`https://github.com/ORGANIZATION/REPOSITORY/settings/hooks/new` where you got to
replace `ORGANIZATION` with your username or organization and `REPOSITORY` with
your repository name, after that follow the steps from the screenshots.

**Payload URL** should be the endpoint where GitHub can access the exporter
through your reverse proxy, or webserver, or whatever you have configured in
front of the exporter, something like `https://exporter.example.com/github`.

**Content type** should be set to `application/json`, but in theory both formats
should be correctly parsed by the exporter.

**Secret** gets the random password you should have prepared when you have
configured the exporter, mentioned above.

**Which events would you like to trigger this webhook** got to be set to
`Let me select individual events` where you just got to check the last item
`Workflow runs`.

After hitting the **Add webhook** button you are ready to receive first webhooks
by GitHub. It should also show that the initial test webhook have been executed
successfully.
