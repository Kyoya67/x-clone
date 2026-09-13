terraform {
  backend "s3" {
    bucket       = "x-clone-terraform-prd"
    key          = "main.tfstate"
    region       = "ap-northeast-1"
    use_lockfile = true
  }
}
