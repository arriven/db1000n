# Hetzner Cloud deployment

## Requirements

- Hetzner Cloud account
- API token (Go to Project - Security - API Tokens) and create a token
- `terraform` (1.0+) installed

## Deploy

To deploy:

```sh
export HCLOUD_TOKEN=<place-api-token-here>
export SSH_PUBLIC_KEY="<place-ssh-public-key-here>"
terraform init
terraform plan -var "hcloud_token=${HCLOUD_TOKEN}" -var "ssh_public_key=${SSH_PUBLIC_KEY}"
terraform apply -var "hcloud_token=${HCLOUD_TOKEN}" -var "ssh_public_key=${SSH_PUBLIC_KEY}"
```

## Destroy

To destroy:

```sh
export HCLOUD_TOKEN=<place-api-token-here>
export SSH_PUBLIC_KEY="<place-ssh-public-key-here>"
terraform destroy -var "hcloud_token=${HCLOUD_TOKEN}" -var "ssh_public_key=${SSH_PUBLIC_KEY}"
```
