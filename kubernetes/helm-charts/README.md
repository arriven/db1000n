# db1000n Helm charts

## If you want to use plain manifests, see [Manifests](/db1000n/advanced-docs/kubernetes/manifests/)

This is a Helm chart for Kubernetes

## Prerequisites

Make sure that you installed `helm` package on your local machine and you have connection to the Kubernetes cluster.

## Install a release

```bash
cd kubernetes/helm-charts/
helm upgrade --install \
    --create-namespace \
    --namespace=db1000n \
    -f values.yaml db1000n .
```

## Destroy a release

```bash
helm uninstall db1000n -n db1000n
```
