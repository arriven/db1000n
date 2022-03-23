# Default AWS provider vars
variable "region" {
  type        = string
  description = "AWS Region"
}

variable "name" {
  type        = string
  description = "name of deployment"
}

variable "arch_ami" {
  type        = string
  description = "architecture of the ami"
  default     = "arm64"
}


variable "instance_type" {
  type        = string
  description = "Instance type"
  default     = "t4g.micro"
}

variable "max_size" {
  type        = number
  description = "Max size of autoscale group"
}

variable "min_size" {
  type        = number
  description = "Min size of autoscale group"
}


# Mixed instances policy part
variable "desired_capacity" {
  type        = number
  description = "number of instances to run"
  default     = 30
}
