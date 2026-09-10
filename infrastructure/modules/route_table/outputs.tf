output "public_id" {
  value       = aws_route_table.public.id
  description = "ID of the public route table."
}

output "app_private_id" {
  value       = aws_route_table.app_private.id
  description = "ID of the application private route table."
}

output "db_private_id" {
  value       = aws_route_table.db_private.id
  description = "ID of the database private route table."
}
