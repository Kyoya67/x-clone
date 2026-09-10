output "address" {
  value       = aws_db_instance.this.address
  description = "RDS hostname used for TLS hostname verification."
}

output "master_user_secret_arn" {
  value       = aws_db_instance.this.master_user_secret[0].secret_arn
  description = "ARN of the RDS-managed administrator secret, not its value."
}
