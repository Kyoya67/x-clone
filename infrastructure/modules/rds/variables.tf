variable "dbadmin_password" {
  type      = string
  sensitive = true
  ephemeral = true
  nullable  = false
}

variable "dbadmin_password_version" {
  type = number
}

variable "dbadmin_references_ready" {
  type        = bool
  default     = false
  description = "Safety gate: enable only after migrating every consumer of the RDS-managed secret."
}

variable "identifier" {
  type        = string
  description = "RDS instance identifier; must be unique within the account and region."
}

variable "engine_version" {
  type        = string
  description = "PostgreSQL 16 version, matching the postgres16 parameter group."
  validation {
    condition     = startswith(var.engine_version, "16.")
    error_message = "This module currently supports PostgreSQL 16 only."
  }
}

variable "instance_class" {
  type        = string
  description = "RDS instance class."
}

variable "multi_az" {
  type        = bool
  description = "Whether to provision a standby in another availability zone."
}

variable "subnet_ids" {
  type        = list(string)
  description = "Database subnet IDs spanning at least two availability zones."
}

variable "security_group_id" {
  type        = string
  description = "Security group allowing PostgreSQL access from application workloads."
}

variable "tags" {
  type    = map(string)
  default = {}
}
