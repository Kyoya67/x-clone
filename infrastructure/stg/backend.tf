terraform {
  backend "s3" {
    bucket = "x-clone-terraform-stg"
    key    = "main.tfstate"
    region = "ap-northeast-1"
  }
}
