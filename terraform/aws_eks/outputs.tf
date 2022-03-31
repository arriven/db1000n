# data sources
output "account_id" {
  description = "AWS Account ID"
  value       = data.aws_caller_identity.current.account_id
}

output "caller_arn" {
  description = "User ARN"
  value       = data.aws_caller_identity.current.arn
}

output "caller_user" {
  value = data.aws_caller_identity.current.user_id
}

# variables
output "region" {
  description = "AWS region"
  value       = var.region
}

output "profile" {
  description = "AWS profile"
  value       = var.profile
}

output "projects" {
  description = "AWS project"
  value       = var.project
}

output "environment" {
  description = "AWS environment"
  value       = terraform.workspace
}