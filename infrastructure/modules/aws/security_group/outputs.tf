output "nat_id" {
  value       = aws_security_group.nat.id
  description = "ID of the NAT instance security group."
}
