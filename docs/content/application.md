---
title: "Application"
date: 2023-10-20T00:00:00+00:00
anchor: "application"
weight: 50
---

Instead of using personal access tokens you can register a GutHub App. Just head
over to `https://github.com/organizations/ORGANIZATION/settings/apps/new` where
you got to replace `ORGANIZATION` with the name of your organization.

Feel free to name the application howerever you like, I have named mine by the
organization and exporter, e.g. `Promhippie Exporter`. For the description you
can write whatever you want or what sounds best for you:

![Application 01](./screenshots/app01.png)

Within the **Identifying and authorizing users** section I have unchecked
everything, at least for me this have worked without any problem so far:

![Application 02](./screenshots/app02.png)

Within the **Post installation** and **Webhook** sections I have also unchecked
everything as this won't be used by the application:

![Application 03](./screenshots/app03.png)

The required permissions have been stripped down to the **Administration**
within **Repository permissions** set to `read-only` and **Administration** as
well as **Self-hosted runners** within **Organization permissions** set to
`read-only`. You are also able to run the exporter with lesser permissions but
than you will loose part of the metrics like the runner or storage related.

Finally I have set the installation to **Only for this account** which seem to
be fine as nobody else got to install the application.

After the creation of the application itself you can already copy the **App ID**
which you will need later to configure the app within the exporter.

Scroll down to **Private keys** and hit the **Generate a private key** button in
order to download the required certificate which you will also need later on.

Since you are still missing the installation ID click on **Install App** on the
left sidebar and install the application for your organization, I have enabled
it far all my repositories. On the page where you are getting redirected to you
got to copy the installation ID from the URL, it's the last numeric part of it.
As an example for me it's something like
`https://github.com/.../installations/43103110` where `43103110` shows the
installation ID you need for the exporter.

## Multiple organizations

If you want to collect metrics from several organizations, or from an
enterprise and its organizations, install the same application on all of them
and leave the installation ID empty. The exporter lists the installations of
the application on startup and uses the matching installation for every
configured enterprise, org and repo owner:

{{< highlight txt >}}
GITHUB_EXPORTER_APP_ID=12345
GITHUB_EXPORTER_PRIVATE_KEY=file:///etc/github_exporter/app.pem
GITHUB_EXPORTER_ENTERPRISE=my-enterprise
GITHUB_EXPORTER_ORG=first-org,second-org
{{< / highlight >}}

Owners the application is not installed on are logged on startup and use the
installation of the first configured owner, enterprises first. The exporter
refuses to start if the application is not installed on any of them.
Installations created later are only picked up after a restart.
