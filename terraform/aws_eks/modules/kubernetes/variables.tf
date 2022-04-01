variable "profile" {
  description = "AWS profile"
}

variable "vpc_id" {
  description = "VPC ID"
}

variable "cluster_name" {
  description = "EKS cluster name"
}

variable "cluster_endpoint" {
  description = "EKS cluster endpoint"
}

variable "cluster_ca_data" {
  description = "EKS cluster certificate authority data"
}

variable "worker_node_iam_role_arn" {
  description = "EKS worker node role ARN"
}

variable "autoscaler_iam_role_arn" {
  description = "EKS cluster autoscaler role ARN"
}