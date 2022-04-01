output "security_group_id" {
  description = "EKS node Security group ID"
  value       = aws_security_group.eks_nodes.id
}

output "worker_node_iam_role_arn" {
  value = aws_iam_role.eks_worker_node.arn
}