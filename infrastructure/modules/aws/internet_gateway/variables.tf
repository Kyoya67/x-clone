variable "name" {
  type        = string
  description = "Prefix used when naming the internet gateway."
}

variable "vpc_id" {
  type        = string
  description = "VPC to which the internet gateway is attached."
}

variable "tags" {
  type        = map(string)
  description = "Tags applied to the internet gateway."
  default     = {}
}
