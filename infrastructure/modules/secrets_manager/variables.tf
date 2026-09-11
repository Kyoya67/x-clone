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

variable "tags" {
  type    = map(string)
  default = {}
}
