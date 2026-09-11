output "address" {
  value       = aws_db_instance.this.address
  description = "RDS hostname used for TLS hostname verification."
}
