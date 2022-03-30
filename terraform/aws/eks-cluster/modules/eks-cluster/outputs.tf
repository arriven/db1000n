output "endpoint" {
  value = aws_eks_cluster.eks_cluster.endpoint
}

output "ca_data" {
  value = aws_eks_cluster.eks_cluster.certificate_authority.0.data
}

output "security_group_id" {
  description = "EKS Control Plane Security group ID"
  value       = aws_security_group.control_plane.id
}

output "autoscaler_iam_role_arn" {
  description = "EKS cluster autoscaler role ARN"
  value       = aws_iam_role.eks_cluster_autoscaler.arn
}