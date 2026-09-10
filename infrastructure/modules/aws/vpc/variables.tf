variable "name" {
  type        = string
  description = "Prefix used when naming the VPC."
}

variable "cidr_block" {
  type        = string
  description = "IPv4 CIDR block assigned to the VPC."
}

variable "tags" {
  type        = map(string)
  description = "Tags applied to the VPC."
  default     = {}
}
