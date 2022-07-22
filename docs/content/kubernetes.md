---
title: "Kubernetes"
date: 2022-07-22T00:00:00+00:00
anchor: "kubernetes"
weight: 20
---

## Kubernetes

Currently we are covering the most famous installation methods on Kubernetes,
you can choose between [Kustomize][kustomize] and [Helm][helm].

### Kustomize

We won't cover the installation of [Kustomize][kustomize] within this guide, to
get it installed and working please read the upstream documentation. After the
installation of [Kustomize][kustomize] you just need to prepare a
`kustomization.yml` wherever you like similar to this:

{{< highlight yaml >}}
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: github-exporter

resources:
  - github.com/promhippie/github_exporter//deploy/kubernetes?ref=master

configMapGenerator:
  - name: github-exporter
    behavior: merge
    literals: []

secretGenerator:
  - name: github-exporter
    behavior: merge
    literals: []
{{< / highlight >}}

After that you can simply execute `kustomize build | kubectl apply -f -` to get
the manifest applied. Generally it's best to use fixed versions of the container
images, this can be done quite easy, you just need to append this block to your
`kustomization.yml` to use this specific version:

{{< highlight yaml >}}
images:
  - name: quay.io/promhippie/github-exporter
    newTag: 1.1.0
{{< / highlight >}}

After applying this manifest the exporter should be directly visible within your
Prometheus instance if you are using the Prometheus Operator as these manifests
are providing a ServiceMonitor.

### Helm

We won't cover the installation of [Helm][helm] within this guide, to get it
installed and working please read the upstream documentation. After the
installation of [Helm][helm] you just need to execute the following commands:

{{< highlight console >}}
helm repo add promhippie https://promhippie.github.io/charts
helm show values promhippie/github-exporter
helm install github-exporter promhippie/github-exporter
{{< / highlight >}}

You can also watch that available values and generally the details of the chart
provided by us within our [chart][chart] repository.

After applying this manifest the exporter should be directly visible within your
Prometheus instance depending on your installation if you enabled the
annotations or the service monitor.

[kustomize]: https://github.com/kubernetes-sigs/kustomize
[helm]: https://helm.sh
[chart]: https://github.com/promhippie/charts/tree/master/charts/github-exporter
