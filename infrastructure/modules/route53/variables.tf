variable "name" {
  type = string
}

variable "tags" {
  type = map(string)
}

variable "api_certificate_domain_validation_options" {
  type = set(object({
    domain_name           = string
    resource_record_name  = string
    resource_record_type  = string
    resource_record_value = string
  }))

  default = []
}

variable "api_alias_record" {
  type = object({
    name     = string
    dns_name = string
    zone_id  = string
  })

  default = null
}

variable "amplify_certificate_dns_record" {
  type = object({
    name  = string
    type  = string
    value = string
  })

  default = null
}

variable "amplify_domain_dns_record" {
  type = object({
    name  = string
    value = string
  })

  default = null
}
