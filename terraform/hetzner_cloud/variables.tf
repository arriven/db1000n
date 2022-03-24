variable "hcloud_token" {
  type      = string
  sensitive = true
}

variable "ssh_public_key" {
  type = string
}

# https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/server#location
variable "location" {
  type    = string
  default = "hel1"
}

variable "instance_count" {
  type    = number
  default = 1
}

# https://www.hetzner.com/cloud
variable "server_type" {
  type    = string
  default = "cx11"
}

variable "os_type" {
  type    = string
  default = "ubuntu-20.04"
}
