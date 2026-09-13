output "public_ids" {
  value       = values(aws_subnet.public)[*].id
  description = "IDs of public subnets."
}

output "private_ids" {
  value       = values(aws_subnet.private)[*].id
  description = "IDs of shared ECS and RDS private subnets."
}

output "private_cidr_blocks" {
  value       = [for subnet in values(aws_subnet.private) : subnet.cidr_block]
  description = "CIDR blocks of shared private subnets."
}
