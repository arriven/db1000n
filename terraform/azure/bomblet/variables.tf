variable "prefix" {
}

variable "region" {
}

variable "bomblet_count" {
}

variable "resource_group_name" {
}

variable "attack_image" {
  default = "ghcr.io/arriven/db1000n:latest"
}

variable "attack_cpu" {
  default = "1"
}

variable "attack_memory" {
  default = "1.5"
}

variable "attack_environment_variables" {
  default = {}
}