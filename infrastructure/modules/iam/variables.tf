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
variable "migration_secret_arn" { type = string }

variable "github_repository" {
  type = string
}

variable "github_oidc_subjects" {
  type = list(string)
}

variable "amplify_app_arn" {
  type = string
}
