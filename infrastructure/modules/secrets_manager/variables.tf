variable "dbadmin" {
  type = object({
    name = string
  })
}

variable "dbadmin_password" {
  type      = string
  sensitive = true
  ephemeral = true
  nullable  = false
}

variable "dbadmin_password_version" {
  type = number
}

variable "app_user" {
  type = object({
    name = string
  })
}

variable "migration_user" {
  type = object({
    name = string
  })
}

variable "auth" {
  type = object({
    name = string
  })
}

variable "auth_client_secret" {
  type      = string
  sensitive = true
  ephemeral = true
  nullable  = false
}

variable "auth_session_secret" {
  type      = string
  sensitive = true
  ephemeral = true
  nullable  = false
}

variable "auth_secret_version" {
  type = number
}

variable "tags" {
  type    = map(string)
  default = {}
}
