variable "backend" {
  description = "通常バックエンドのタスク設定。"
  type = object({
    family              = string
    image               = string
    execution_role_arn  = string
    task_role_arn       = string
    database_host       = string
    database_secret_arn = string
    auth_secret_arn     = string
    auth = object({
      issuer         = string
      authorize_url  = string
      token_url      = string
      client_id      = string
      redirect_url   = string
      post_login_url = string
      cookie_domain  = string
      cookie_secure  = string
      cookie_prefix  = string
    })
    log_group_name = string
  })
}

variable "migration" {
  description = "単発マイグレーションのタスク設定。"
  type = object({
    family               = string
    image                = string
    execution_role_arn   = string
    task_role_arn        = string
    database_host        = string
    migration_secret_arn = string
    log_group_name       = string
  })
}

variable "region" { type = string }

variable "tags" {
  type    = map(string)
  default = {}
}
