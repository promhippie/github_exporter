---
title: "Building"
date: 2022-07-20T00:00:00+00:00
anchor: "building"
weight: 20
---

As this project is built with Go you need to install Go first. If you are not
familiar with [Nix][nix] it is up to you to have a working environment for Go
(>= 1.24.0) as the setup won't we covered within this guide. Please follow the
official install instructions for [Go][golang]. Beside that we are using
[go-task][gotask] to define all commands to build this project.

{{< highlight txt >}}
git clone https://github.com/promhippie/github_exporter.git
cd github_exporter/

task generate build
./bin/github_exporter -h
{{< / highlight >}}

[nix]: https://nixos.org/
[golang]: http://golang.org/doc/install.html
[gotask]: https://taskfile.dev/installation/
