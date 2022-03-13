# GCP + ExpressVPN deployment

## Requirements

- [GCP account](http://console.cloud.google.com)
- Subscription on [expressvpn.com](https://www.expressvpn.com) (get the activation code)
- `terraform` installed

## Init

To init Terraform run:

```sh
terraform init
```

Need to create terraform/gcp_expressvpn/terraform.tfvars file and set two variables values tou yours

```sh
project_id     = "google-project-id"
expressvpn_key = "expressvpn-activation-code"
```

Other vars can be overwritten in this file id needed.

## Deploy

To deploy run:

```sh
terraform apply
```

## Destroy

To destroy infrastructure use commands:

```sh
terraform destroy
```
