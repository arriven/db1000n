# Deployment

## Requirements

    - AWS account
    - terraform installed

## Deploy

to deploy run

```terraform
terraform apply -var-file="ireland.tfvars"
```

you can create new tfvars files for different regions and accounts  
to swich between regions you can sue `terraform workspace` command
for example:

```terraform
terraform init
terraform workspace new eu
terraform apply -var-file="ireland.tfvars"
terraform workspace new us
terraform apply -var-file="useast.tfvars"
```

to destroy infrastructure you can use next commands

```terraform
terraform workspace select eu
terraform destroy -var-file="ireland.tfvars"
terraform workspace select us
terraform destroy -var-file="useast.tfvars"
```
