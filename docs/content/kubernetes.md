---
title: "Kubernetes"
date: 2022-03-09T00:00:00+00:00
anchor: "kubernetes"
weight: 20
---

## Kubernetes

So far we got the deployment via [Kustomize](https://github.com/kubernetes-sigs/kustomize) to get this exporter working on Kubernetes. We are already working on a [Helm]() chart to offer more options, dependening on your preferences.

### Kustomize

We won't cover the installation of [Kustomize](https://github.com/kubernetes-sigs/kustomize) or encryption tooling like [KSOPS](https://github.com/viaduct-ai/kustomize-sops) within this guide, to get it installed and working please consult the documentation of these projects. After the installation of [Kustomize](https://github.com/kubernetes-sigs/kustomize) you just need to prepare a `kustomization.yml` wherever you like:

{{< highlight yaml >}}
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: github-exporter

resources:
  - github.com/promhippie/github_exporter?ref=master

configMapGenerator:
  - name: github-exporter
    behavior: merge
    literals:
      - GITHUB_EXPORTER_ORGS=promhippie
      - GITHUB_EXPORTER_REPOS=promhippie/*

secretGenerator:
  - name: github-exporter
    behavior: merge
    literals:
      - GITHUB_EXPORTER_TOKEN=bldyecdtysdahs76ygtbw51w3oeo6a4cvjwoitmb
{{< / highlight >}}

After that you can simply execute `kustomize build | kubectl apply -f -` to get the manifest applied. Generally it's best to use fixed versions of Docker images, this can be done quite easy, you just need to append this block to your `kustomization.yml` to use this specific version:

{{< highlight yaml >}}
images:
  - name: quay.io/promhippie/github-exporter
    newTag: 1.1.0
{{< / highlight >}}

After applying this manifest the exporter should be directly visible within your Prometheus instance if you are using the Prometheus Operator as these manifests are providing a ServiceMonitor.
