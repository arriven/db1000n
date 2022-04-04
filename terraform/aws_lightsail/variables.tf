variable "app" {
  type    = string
  default = "db1000n"
}

variable "image" {
  type    = string
  default = "ghcr.io/arriven/db1000n:latest"
}

variable "scale" {
  type    = number
  default = 1
}

# https://lightsail.aws.amazon.com/ls/docs/en_us/articles/amazon-lightsail-creating-container-services#create-container-service-capacity
variable "power" {
  type    = string
  default = "medium"
}

variable "region_name" {
  type    = string
  default = "eu-central-1"
}
