# DO terraform module [db1000n](https://github.com/Arriven/db1000n)

This module creates a number of servers in each of provided region. If you set `count = 2` and regions = ["nyc1", "nyc2", "nyc3"] this will create six servers total. Two servers in each of regions.

**Requriment:**
- Digital Ocean account
- API key
- Already present SSH key in DO account
- Terraform

![](https://raw.githubusercontent.com/dddbbbsss/terraform_db1000n/main/img/SCR-20220306-sq0.png)
![](https://raw.githubusercontent.com/dddbbbsss/terraform_db1000n/main/img/SCR-20220306-sqw.png)

## ADD API key
`./module/variables.tf`

```
variable "do_token" {
  type    = string
  default = "your_API_key"
}
```


## How to add SSH key
Settings -> Security -> Add SSH key

remeber key's name and add it into: `./03-main.tf` string `digitalocean_ssh_key =` 

```
module "db1000n" {
  source               = "./module"
  count                = 2
  regions              = ["nyc1", "nyc3", "sfo3", "ams3", "sgp1", "lon1", "fra1", "tor1", "blr1"]
  name                 = "db00-${count.index}"
  digitalocean_tag     = "stop-sites"
  image_name           = "ubuntu-20-04-x64"
  size                 = "s-1vcpu-1gb"
  ipv6                 = true
  backups              = false
  monitoring           = true
  droplet_agent        = true
  tags                 = "stop-sites"
  digitalocean_ssh_key = "SSH_key_name"
}
```

`count =` - it's number of droplets creates in each of `regions` 

Version - `db1000n_version = "v0.5.20"` actual version you may get [there](https://github.com/Arriven/db1000n/releases)


## Also means 
you are using ssh keys with name `~/.ssh/id_rsa.pub` if not, change it in `./module/variables.tf` these variables `variable "pub_key"` and `variable "pvt_key"`

Run `terraform init` If I didn't miss anything, you will not get an error message.
Then `terraform plan` and `terraform apply`
delete `terraform destroy -auto-approve`
	
