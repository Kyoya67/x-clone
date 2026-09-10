terraform {
  required_version = "~> 1.14.1"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.34.0"
    }
  }
}

provider "aws" {
  region  = "ap-northeast-1"
  profile = "x-clone-terraform-stg"

  default_tags {
    tags = {
      Env = "stg"
    }
  }
}
