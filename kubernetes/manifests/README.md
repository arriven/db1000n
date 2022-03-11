# Kubernetes manifests to install

## If you use Helm, see our [Helm Chart](/db1000n/advanced-docs/kubernetes/helm-charts/)

There are two ways to deploy it with plain manifests:

- using Deployment
- using DaemonSet

## Deployment

Install:

```bash
cd kubernetes/manifests/
kubectl apply -f deployment.yaml
kubectl get po -n db1000n
```

Scale:

```bash
kubectl scale deployment/db1000n --replicas=10 -n db1000n
```

Destroy:

```bash
kubectl delete deploy db1000n -n db1000n
```

## DaemonSet

Get and label nodes where you need to run `db1000n`.
There should be nodes at least with 2CPU and 2GB of RAM, CPU resources in priority for `db1000n`:

```bash
kubectl get nodes
```

Select nodes where you want to run `db1000n` from the output and label them:

```bash
kubectl label nodes <YOUR_UNIQUE_NODE_NAME> db1000n=true
```

Install the DaemonSet:

```bash
kubectl apply -f daemonset.yaml
```

Destroy:

```bash
kubectl delete daemonset db1000n -n db1000n
```

???+ info "How it works?"

    DaemonSet will create one `db1000n` pod on each node that labeled as `db1000n=true`.
    It coule be useful in large cluster types that can be autoscaled horizontally, for example, GKE standard k8s cluster from the free tier purposes.
