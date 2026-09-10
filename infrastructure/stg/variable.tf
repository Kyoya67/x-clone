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
