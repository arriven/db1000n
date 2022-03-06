# db1000n Helm chart

*If you want to use plain manifess, see manifests [here](../manifests/)*

This is a Helm chart for Kubernetes

## Prerequisites

Make sure that you installed `helm` package on your local machine and you have connection to the Kubernetes cluster.

## Install chart

```bash
cd kubernetes/helm-charts/
helm upgrade --install \
    --create-namespace \
    --namespace=db1000n \
    -f values.yaml db1000n .
```

## Destroy chart

```bash
helm uninstall db1000n -n db1000n
```
