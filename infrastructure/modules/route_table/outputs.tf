output "public_id" {
  value       = aws_route_table.public.id
  description = "ID of the public route table."
}

output "private_id" {
  value       = aws_route_table.private.id
  description = "ID of the shared private route table."
}
