# Vultr deployment

## Requirements

* Vultr account
* API token
* `terraform`

## Deploy

```bash
export VULTR_API_KEY="Your Vultr API Key"
terraform init
terraform plan -var "key=<path_to_ssh_key>" -var "num_inst=<number of instances to create>"
terraform apply -var "key=<path_to_ssh_key>" -var "num_inst=<number of instances to create>"
```

## Destroy

To delete all the resources that were created run

```bash
terraform destroy
```

## Tips

Deploy script installs vnstat util that is useful for monitoring server network performance.
Example, get network statistics for the last 5 hours:

```bash
ssh root@ip vnstat -h 5
```
