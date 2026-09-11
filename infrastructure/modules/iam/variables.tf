variable "repository_arn" {
  type = string
}

variable "log_group_arn" {
  type = string
}

variable "database_secret_arn" {
  type = string
}

variable "tags" {
  type    = map(string)
  default = {}
}

variable "migration_repository_arn" { type = string }
variable "migration_log_group_arn" { type = string }
variable "admin_secret_arn" { type = string }
