output "public_ids" {
  value       = values(aws_subnet.public)[*].id
  description = "IDs of public subnets."
}

output "app_private_ids" {
  value       = values(aws_subnet.app_private)[*].id
  description = "IDs of private application subnets."
}

output "app_private_cidr_blocks" {
  value       = [for subnet in values(aws_subnet.app_private) : subnet.cidr_block]
  description = "CIDR blocks of private application subnets."
}

output "db_private_ids" {
  value       = values(aws_subnet.db_private)[*].id
  description = "IDs of isolated database subnets."
}
