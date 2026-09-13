variable "name" { type = string }
variable "domain_prefix" { type = string }
variable "callback_urls" { type = list(string) }
variable "logout_urls" { type = list(string) }
variable "google_client_id" { type = string }
variable "google_client_secret" {
  type      = string
  sensitive = true
}
variable "tags" {
  type    = map(string)
  default = {}
}

