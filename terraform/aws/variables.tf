# Default AWS provider vars
variable "region" {
  type        = string
  description = "AWS Region"
}

variable "name" {
  type        = string
  description = "name of deployment"
}

variable "instance_type" {
  type        = string
  description = "Instance type"
  default     = "t2.micro"
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
