# common
variable "region" {
  description = "AWS region"
  default     = "us-east-1"
}

variable "profile" {
  description = "AWS profile"
  default     = "default"
}

variable "project" {
  description = "A list of projects to deploy"
  default     = "db1000n"
}

# ssh key
variable "key_name" {
  description = "SSH key name"
  default     = "db1000n"
}

variable "public_key" {
  description = "SSH public key"
}

# eks
variable "eks_node_instance_type" {
  description = "EC2 instance type for EKS nodes"
  default     = "t3.medium"
}

variable "eks_node_desired_capacity" {
  description = "The number of Amazon EC2 instances that should be running in the auto scale group"
  default     = "3"
}

variable "eks_node_max_size" {
  description = "The maximum size of the auto scale group"
  default     = "4"
}

variable "eks_node_min_size" {
  description = "The minimum size of the auto scale group"
  default     = "2"
}

variable "tags" {
  description = "A map of tags to add to VPC"
  type        = map(string)
  default     = {}
}