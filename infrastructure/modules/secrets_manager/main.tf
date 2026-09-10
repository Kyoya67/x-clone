# 接続URLの値は登録しない。Terraform Stateに認証情報を保存しないため。
resource "aws_secretsmanager_secret" "this" {
  name                    = var.name
  description             = "Application database URL; populate before starting ECS tasks"
  recovery_window_in_days = 7
  tags                    = var.tags
}
