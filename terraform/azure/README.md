# Deploy via Azure

## Prerequisites

- Install terraform
- Register a new Azure account by providing a valid credit card and get 200$ free credits.
- [Prepare environment for Azure Provider](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs)
- [The easiest option for auth is Azure CLI](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/guides/azure_cli)

## Deployment

The composition creates container instances in 6 different regions for a broader attack. If you want to make a different setup, just alter modules in the `main.tf`.

Create a new `terraform.tfvars` file in the folder, if you want to change the default configuration of the farm:

- `bomblet_count=10` - can be used for custom number of containers per region
- `attack_commands=["/usr/src/app/main","-c=https://link_to_your_config_file"]`

`terraform init` - to restore all dependencies.

`terraform apply -auto-approve` - to provision the attack farm.

## Collecting logs from the containers

The container instances are provisioned without public IP addresses to make the setup more cost effective.
If you deploy more than one container per region, play with the `-01` suffix to get logs from the correct instance.

- Logs from North Europe region:

```sh
az container logs --resource-group attack-rg --name attack-northeurope-01 --container-name main
```

- Logs from West Europe region:

```sh
az container logs --resource-group attack-rg --name attack-westeurope-01 --container-name main
```

- Logs from Canada Central region:

```sh
az container logs --resource-group attack-rg --name attack-canadacentral-01 --container-name main
```

- Logs from UAE North region:

```sh
az container logs --resource-group attack-rg --name attack-uaenorth-01 --container-name main
```

- Logs from Central US region:

```sh
az container logs --resource-group attack-rg --name attack-centralus-01 --container-name main
```

- Logs from East Asia region:

```sh
az container logs --resource-group attack-rg --name attack-eastasia-01 --container-name main
```

## Cleanup

```sh
terraform destroy
```
