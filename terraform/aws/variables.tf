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

variable "zones" {
  type        = number
  description = "number of availability zones"
  default     = 2
}

# if you have multiple aws accounts you are managing with
# terraform eg with aws-vault, specify your auth profile here.
# leave null to use default profile
variable "profile" {
  type        = string
  description = "aws auth profile"
  default     = null
}

variable "allow_ssh" {
  type        = bool
  description = "allow port 22 access to proxy and db1000n instances"
  default     = true
}

# Optional. I use this to set ec2-user's password, enabling serial port
# access to ec2 instances via the AWS console, even for instances in private
# networks. IMHO this is more secure than exposing port 22 to the outside world
# example: "usermod --password <encrypted password> ec2-user"
variable "extra_startup_script" {
  type        = string
  description = "commands to append to instance startup script"
  default     = ""
}

variable "enable_tor_proxy" {
  type        = bool
  description = "create tor proxy for outbound connections"
  default     = false
}