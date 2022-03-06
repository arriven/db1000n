# Heroku deployment

## Requirements

- Heroku account
- API token (Go to Account settings - API Key) and reveal API key
- `terraform` (1.0+) installed

## Deploy

To deploy run:

```terraform
export EMAIL=<place-email-here>
export API_KEY=<place-api-key-here>
terraform init
terraform plan -var "email=${EMAIL}" -var "api_key=${API_KEY}"
terraform apply -var "email=${EMAIL}" -var "api_key=${API_KEY}"
```

Go to [apps list](https://dashboard.heroku.com/apps) and ensure that application successfully deployed.
You can check logs for application with Heroku CLI: https://devcenter.heroku.com/articles/logging#view-logs

## Destroy

To destroy infrastructure use commands:

```terraform
export EMAIL=<place-email-here>
export API_KEY=<place-api-key-here>
terraform destroy -var "email=${EMAIL}" -var "api_key=${API_KEY}"
```
