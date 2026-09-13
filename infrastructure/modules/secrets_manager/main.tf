# dbadminの値はwrite-onlyで登録し、Stateには保存しない。
resource "aws_secretsmanager_secret" "dbadmin" {
  name                    = var.dbadmin.name
  description             = "dbadmin: administrator credentials for RDS"
  recovery_window_in_days = 7
  tags                    = var.tags
}

resource "aws_secretsmanager_secret_version" "dbadmin" {
  secret_id = aws_secretsmanager_secret.dbadmin.id
  secret_string_wo = jsonencode({
    username = "dbadmin"
    password = var.dbadmin_password
  })
  secret_string_wo_version = var.dbadmin_password_version
}

# app_user・migration_userの値はGoコマンドから登録する。
resource "aws_secretsmanager_secret" "app_user" {
  name                    = var.app_user.name
  description             = "app_user: username/password JSON for the backend application"
  recovery_window_in_days = 7
  tags                    = var.tags
}

resource "aws_secretsmanager_secret" "migration_user" {
  name                    = var.migration_user.name
  description             = "migration_user: username/password JSON for schema migrations"
  recovery_window_in_days = 7
  tags                    = var.tags
}

resource "aws_secretsmanager_secret" "auth" {
  name                    = var.auth.name
  description             = "auth: OIDC client secret and session signing secret"
  recovery_window_in_days = 7
  tags                    = var.tags
}

resource "aws_secretsmanager_secret_version" "auth" {
  secret_id = aws_secretsmanager_secret.auth.id
  secret_string_wo = jsonencode({
    client_secret  = var.auth_client_secret
    session_secret = var.auth_session_secret
  })
  secret_string_wo_version = var.auth_secret_version
}
