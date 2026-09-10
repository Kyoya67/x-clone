variable "name" {
  type        = string
  description = "Prefix used when naming the NAT instance."
}

variable "public_subnet_id" {
  type        = string
  description = "Public subnet in which the NAT instance is launched."
}

variable "nat_security_group_id" {
  type        = string
  description = "Security group attached to the NAT instance."
}

variable "nat_instance_type" {
  type        = string
  description = "EC2 instance type for the NAT instance."
  default     = "t3.micro"
}

variable "tags" {
  type        = map(string)
  description = "Tags applied to the NAT instance."
  default     = {}
}
