variable "backend" {
  description = "通常バックエンドのタスク設定。"
  type = object({
    family              = string
    image               = string
    execution_role_arn  = string
    task_role_arn       = string
    database_host       = string
    database_secret_arn = string
    log_group_name      = string
  })
}

variable "migration" {
  description = "単発マイグレーションのタスク設定。"
  type = object({
    family             = string
    image              = string
    execution_role_arn = string
    task_role_arn      = string
    database_host      = string
    admin_secret_arn   = string
    log_group_name     = string
  })
}

variable "region" { type = string }

variable "tags" {
  type    = map(string)
  default = {}
}
