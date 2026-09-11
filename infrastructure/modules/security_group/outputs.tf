output "nat_id" {
  value       = aws_security_group.nat.id
  description = "ID of the NAT instance security group."
}

output "backend_id" {
  value       = aws_security_group.backend.id
  description = "Security group for backend ECS tasks and migration tasks."
}

output "db_id" {
  value       = aws_security_group.db.id
  description = "Security group for RDS."
}
