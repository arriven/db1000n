# AWS deployment

## Requirements

- AWS account
- `terraform` installed

## Deploy

To deploy run:

```terraform
terraform apply -var-file="ireland.tfvars"
```

You can create new `*.tfvars` files for different regions and accounts.
To swich between regions you can use `terraform workspace` command.

For example:

```terraform
terraform init
terraform workspace new eu
terraform apply -var-file="ireland.tfvars"
terraform workspace new us
terraform apply -var-file="useast.tfvars"
```

## Destroy

To destroy infrastructure use commands:

```terraform
terraform workspace select eu
terraform destroy -var-file="ireland.tfvars"
terraform workspace select us
terraform destroy -var-file="useast.tfvars"
```
