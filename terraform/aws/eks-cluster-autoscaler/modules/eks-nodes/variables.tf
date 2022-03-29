variable "autoscale_group_name" {
  description = "Name to use for the auto scale group"
}

variable "cluster_name" {
  description = "EKS cluster name"
}

variable "cluster_version" {
  description = "EKS cluster version"
}

variable "cluster_endpoint" {
  description = "EKS cluster endpoint"
}

variable "cluster_ca_data" {
  description = "EKS cluster certificate authority"
}

variable "vpc_id" {
  description = "VPC ID"
}

variable "source_security_groups" {
  description = "A list of source security groups which can connect to the EKS nodes"
}

variable "subnets" {
  description = "A list of subnets to place the EKS nodes"
  type        = list(string)
}

variable "key_name" {
  description = "SSH key name"
}

variable "public_key" {
  description = "SSH public key"
}

variable "ami_id" {
  description = "The AMI from which to launch the instances"
  default     = ""
}

variable "instance_type" {
  description = "The type of the instance"
}

variable "device_name" {
  description = "The name of the device to mount."
  default     = "/dev/xvda"
}

variable "volume_type" {
  description = "The type of volume. Can be `standard`, `gp2`, or `io1`."
  default     = "gp2"
}

variable "volume_size" {
  description = "The size of the volume in gigabytes."
  default     = "20"
}

variable "encrypted" {
  description = "Enables EBS encryption on the volume."
  default     = true
}

variable "desired_capacity" {
  description = "The number of Amazon EC2 instances that should be running in the auto scale group"
}

variable "max_size" {
  description = "The maximum size of the auto scale group"
}

variable "min_size" {
  description = "The minimum size of the auto scale group"
}

variable "on_demand_base_capacity" {
  description = "Auto Scalling Group value for desired capacity for instance lifecycle type on-demand of bastion hosts."
  default     = 0
}

variable "use_spot_instances" {
  description = "Use spot instances or on-demand"
  default     = false
}

variable "spot_overrides" {
  description = "Instance type overrides. Only applicable with spot instances"
  type = list(object({
    instance_type     = string
    weighted_capacity = number
  }))
  default = []
}

variable "termination_policies" {
  type        = list(string)
  description = "A list of policies to decide how the instances in the auto scale group should be terminated. The allowed values are OldestInstance, NewestInstance, OldestLaunchConfiguration, ClosestToNextInstanceHour, OldestLaunchTemplate, AllocationStrategy."
  default     = ["OldestInstance"]
}