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

variable "migration_image_tag" {
  type        = string
  description = "First six characters of the Git commit used to build the migration image."
  validation {
    condition     = can(regex("^[0-9a-f]{6}$", var.migration_image_tag))
    error_message = "migration_image_tag must be a six-character lowercase Git commit hash."
  }
}
