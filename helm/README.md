# db1000n Helm chart

This is a Helm chart for Kubernetes

## Quick start

### Prerequisites

Make sure that you installed `helm` package on your local machine and you have connection to the Kubernetes cluster.

### Install chart

```bash
cd helm/
helm upgrade --install \
    -n db1000n \
    -f values.yaml db1000n .
```
