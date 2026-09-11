locals {
  env        = "stg"
  account_id = "089244387218"
  region     = "ap-northeast-1"

  common_tags = {
    Project   = "x-clone"
    ManagedBy = "Terraform"
    Env       = local.env
  }
}

variable "dbadmin_password" {
  type        = string
  sensitive   = true
  ephemeral   = true
  nullable    = false
  description = "Administrator password supplied securely at runtime; never place it in committed tfvars."

  validation {
    # RDS for PostgreSQL: 8〜128文字の表示可能ASCII。空白・/・\"・@は不可。
    condition = (
      can(regex("^[!-~]{8,128}$", var.dbadmin_password)) &&
      !can(regex("[/\"@]", var.dbadmin_password))
    )
    error_message = "Use 8 to 128 printable ASCII characters, excluding spaces, /, \", and @."
  }
}

variable "dbadmin_password_version" {
  type        = number
  default     = 1
  description = "Increment when changing the administrator password; shared by RDS and Secrets Manager."

  validation {
    condition     = var.dbadmin_password_version >= 1 && floor(var.dbadmin_password_version) == var.dbadmin_password_version
    error_message = "Password version must be a positive integer."
  }
}
