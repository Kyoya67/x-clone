output "nat_id" {
  value       = aws_security_group.nat.id
  description = "ID of the NAT instance security group."
}

output "api_alb_id" {
  value = aws_security_group.api_alb.id
}

output "api_id" {
  value       = aws_security_group.api.id
  description = "Security group for API ECS tasks."
}

output "db_id" {
  value       = aws_security_group.db.id
  description = "Security group for RDS."
}
