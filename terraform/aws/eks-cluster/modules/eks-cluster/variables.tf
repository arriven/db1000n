variable "vpc_id" {
  description = "VPC ID"
}

variable "cluster_name" {
  description = "EKS cluster name"
}

variable "cluster_version" {
  description = "EKS cluster version"
}

variable "source_security_groups" {
  description = "A list of source security groups which can connect to the EKS cluster"
}

variable "subnets" {
  description = "A list of subnets to place the EKS cluster"
  type        = list(string)
}

