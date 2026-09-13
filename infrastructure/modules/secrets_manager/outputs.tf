output "dbadmin_secret_arn" {
  value      = aws_secretsmanager_secret.dbadmin.arn
  depends_on = [aws_secretsmanager_secret_version.dbadmin]
}

output "app_user_secret_arn" {
  value = aws_secretsmanager_secret.app_user.arn
}

output "migration_user_secret_arn" {
  value = aws_secretsmanager_secret.migration_user.arn
}

output "auth_secret_arn" {
  value      = aws_secretsmanager_secret.auth.arn
  depends_on = [aws_secretsmanager_secret_version.auth]
}
