variable "pub_key" {
  default = "~/.ssh/id_rsa.pub"
}

variable "pvt_key" {
  default = "~/.ssh/id_rsa"
}

variable "regions" {
  type = list(string)
}

variable "name" {
  type = string
}

variable "digitalocean_tag" {
  type = string
}

variable "size" {
  type = string
}

variable "ipv6" {
  type = string
}

variable "backups" {
  type = string
}

variable "monitoring" {
  type = string
}

variable "droplet_agent" {
  type = string
}

variable "image_name" {
  type = string
}

variable "tags" {
  type = string
}

variable "digitalocean_ssh_key" {
  type = string
}

variable "db1000n_version" {
  type = string
}
