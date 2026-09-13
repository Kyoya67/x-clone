variable "domain_name" {
  type = string
}

variable "tags" {
  type = map(string)
}

variable "validation_record_fqdns" {
  type    = list(string)
  default = []
}
