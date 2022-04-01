# AWS EKS deployment

## Description

This implementation allows you to create entire AWS infrastructure from scratch
and provides Kubernetes cluster (EKS) to deploy **db1000n** project.

## Prerequisites

- AWS account with **AdministratorAccess** permissions
- OS Linux or Windows
- [AWS CLI](https://docs.aws.amazon.com/cli/v1/userguide/cli-chap-install.html)
- [Terraform](https://learn.hashicorp.com/tutorials/terraform/install-cli)
- [Helm](https://helm.sh/docs/intro/install/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)

## Configure AWS profile

The following example shows sample values:

```bash
$ aws configure
AWS Access Key ID [None]: AKIAIOSFODNN7EXAMPLE
AWS Secret Access Key [None]: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
Default region name [None]: us-west-2
Default output format [None]: json
```

## Deployment

### Deploy infrastructure

```bash
cd db1000n/terraform/aws/eks-cluster/
terraform init
terraform plan
terraform apply
```

**NOTE:** You can create multilpe `*.tfvars` configuration files with various variables, regions and AWS accounts
using `terraform workspace` command:

```bash
cd db1000n/terraform/aws/eks-cluster/
terraform init
terraform workspace new $your_workspace
terraform plan -var-file $your_file.tfvars
terraform apply -var-file $your_file.tfvars
```

### Update kubeconfig

```bash
aws --profile $your_aws_profile eks update-kubeconfig --name $your_eks_cluster_name
```

### Connect to EKS cluster

```bash
$ kubectl get nodes
NAME                                          STATUS   ROLES    AGE    VERSION
ip-xxx-xxx-x-xx.us-east-1.compute.internal    Ready    <none>   107m   v1.21.5-eks-9017834
ip-xxx-xxx-x-xx.us-east-1.compute.internal    Ready    <none>   107m   v1.21.5-eks-9017834
ip-xxx-xxx-x-xx.us-east-1.compute.internal    Ready    <none>   107m   v1.21.5-eks-9017834
```

### Install application

```bash
$ cd db1000n/kubernetes/helm-charts/
$ helm upgrade --install \
    --create-namespace \
    --namespace=db1000n \
    -f values.yaml db1000n .
```

### Check installation

```bash
$ kubectl -n db1000n get pods
NAME                       READY   STATUS    RESTARTS   AGE
db1000n-54d8744b54-8hffr   1/1     Running   0          2m10s
db1000n-54d8744b54-8vml4   1/1     Running   0          2m10s
db1000n-54d8744b54-9stzv   1/1     Running   0          2m10s
```

## Deletion

### Delete application

```bash
helm uninstall db1000n -n db1000n
```

### Delete infrastructure

```bash
terraform destroy
```
