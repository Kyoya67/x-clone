variable "name" {
  type = string
}

variable "branch_name" {
  type = string
}

variable "tags" {
  type = map(string)
}

variable "api_url" {
  description = "API転送先のHTTPSオリジン。パスや末尾のスラッシュを含めない。"
  type        = string

  validation {
    condition     = can(regex("^https://[A-Za-z0-9.-]+$", var.api_url))
    error_message = "api_url must be an HTTPS origin without a path or trailing slash."
  }
}
