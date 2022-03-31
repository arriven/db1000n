variable "name" {
  description = "VPC name"
}

variable "tags" {
  description = "A map of tags to add to VPC"
}

variable "vpc_cidr_block" {
  description = "Base CIDR block which is divided into subnet CIDR blocks (e.g. `10.0.0.0/16`)"
  default     = "10.0.0.0/16"
}

variable "amount_az" {
  description = "Desired Availability Zones (must be greater than 0). Number of available zones depends on region."
  default     = "2"
}