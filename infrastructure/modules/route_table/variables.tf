variable "vpc_id" {
  type        = string
  description = "VPC in which route tables are created."
}

variable "internet_gateway_id" {
  type        = string
  description = "Internet gateway used by public subnets."
}

variable "nat_network_interface_id" {
  type        = string
  description = "NAT instance network interface used by application private subnets."
}

variable "public_subnet_ids" {
  type        = list(string)
  description = "Public subnets associated with the public route table."
}

variable "app_private_subnet_ids" {
  type        = list(string)
  description = "Application private subnets associated with the NAT route table."
}

variable "db_private_subnet_ids" {
  type        = list(string)
  description = "Database private subnets associated with the isolated route table."
}

variable "tags" {
  type        = map(string)
  description = "Tags applied to all route tables."
  default     = {}
}
