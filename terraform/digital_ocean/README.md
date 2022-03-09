# Digital Ocean deployment

## Requirements

- Digital Ocean account
- API token (Go to API - Personal access tokens) and generate Personal access token (with write permissions)
- `terraform` (1.0+) installed

## Deploy

To deploy run:

```sh
export DO_TOKEN=<place-api-token-here>
terraform init
terraform plan -var "do_token=${DO_TOKEN}"
terraform apply -var "do_token=${DO_TOKEN}"
```

After deployment (usually takes 5-10 mins) go to [Apps List](https://cloud.digitalocean.com/apps), find an app with name `db1000n` and chek Runtime Logs.

## Destroy

To destroy infrastructure use commands:

```sh
export DO_TOKEN=<place-api-token-here>
terraform destroy -var "do_token=${DO_TOKEN}"
```
