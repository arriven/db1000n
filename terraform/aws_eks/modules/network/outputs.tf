output "vpc_id" {
  description = "VPC ID"
  value       = aws_vpc.vpc.id
}

output "availability_zones" {
  description = "List of available Availability Zones for selected region"
  value       = data.aws_availability_zones.available.names
}

output "public_subnet_ids" {
  description = "Public subnet ID"
  value       = aws_subnet.public.*.id
}

output "private_subnet_ids" {
  description = "Private subnet ID"
  value       = aws_subnet.private.*.id
}