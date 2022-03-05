# Attack via Azure

## Prerequisites

- Install terraform
- [Prepare evironment for Azure Provider](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs)
    - [The easiest option for auth is AzureCLI](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/guides/azure_cli)

## Attack

The composition creates contaner instances in 6 different regions for a more broad attack. If you want to make different setup, just alter modules in the main.tf.

- Create a new terraform.tfvars file in the folder, if you want change the default configuration of the farm:
    - bomblet_count=10   - can be used for custom number of containers per region.

- ```terrafrom init``` - to restore all dependencies.

- ```terraform apply -auto-approve``` - to provision the attack farm. 


## Collecting logs from the containers

The container instances are provisioned without public IP adresses to make the setup more cost effective. If you deploy more than one container per region, play with the -01 sufix to get logs from the correct instance.

- Logs from North Europe region

```
az container logs --resource-group attack-rg --name attack-northeurope-01 --container-name main
```

- Logs from West Europe region

```
az container logs --resource-group attack-rg --name attack-westeurope-01 --container-name main
```

- Logs from Canada Central region 

```
az container logs --resource-group attack-rg --name attack-canadacentral-01 --container-name main
```

- Logs from UAE North region

```
az container logs --resource-group attack-rg --name attack-uaenorth-01 --container-name main
```

- Logs from Central US region

```
az container logs --resource-group attack-rg --name attack-centralus-01 --container-name main
```

- Logs from East Asia region

```
az container logs --resource-group attack-rg --name attack-eastasia-01 --container-name main
```

## Cleanup

```terraform destroy```