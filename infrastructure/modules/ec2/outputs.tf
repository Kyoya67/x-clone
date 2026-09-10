output "nat_id" {
  value       = aws_instance.nat.id
  description = "ID of the NAT instance."
}

output "nat_network_interface_id" {
  value       = aws_instance.nat.primary_network_interface_id
  description = "Primary network interface ID of the NAT instance."
}
