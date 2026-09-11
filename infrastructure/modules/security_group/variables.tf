variable "vpc_id" {
  type        = string
  description = "VPC in which the security group is created."
}

variable "app_private_cidr_blocks" {
  type        = list(string)
  description = "Private application subnet CIDR blocks allowed to use the NAT instance."
}

variable "tags" {
  type        = map(string)
  description = "Tags applied to the security group."
  default     = {}
}
