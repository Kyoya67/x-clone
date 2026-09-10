variable "repository_arn" {
  type = string
}

variable "log_group_arn" {
  type = string
}

variable "database_url_secret_arn" {
  type = string
}

variable "tags" {
  type    = map(string)
  default = {}
}
