locals {
  env        = "prd"
  account_id = "517037063215"
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

variable "google_client_id" {
  type        = string
  nullable    = false
  description = "Google OAuth client ID for Cognito Hosted UI."
}

variable "google_client_secret" {
  type        = string
  sensitive   = true
  nullable    = false
  description = "Google OAuth client secret for Cognito Hosted UI."
}

variable "auth_session_secret" {
  type        = string
  sensitive   = true
  ephemeral   = true
  nullable    = false
  description = "Secret used by the backend to sign httpOnly session cookies."
}

variable "auth_secret_version" {
  type        = number
  default     = 1
  description = "Increment when changing auth secrets."

  validation {
    condition     = var.auth_secret_version >= 1 && floor(var.auth_secret_version) == var.auth_secret_version
    error_message = "Auth secret version must be a positive integer."
  }
}
