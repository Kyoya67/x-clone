variable "name" {
  type        = string
  description = "Prefix used when naming subnets."
}

variable "vpc_id" {
  type        = string
  description = "VPC in which subnets are created."
}

variable "availability_zones" {
  type        = list(string)
  description = "Availability Zones in which each subnet tier is created."

  validation {
    condition     = length(var.availability_zones) >= 2
    error_message = "At least two Availability Zones are required."
  }
}

variable "public_cidr_blocks" {
  type        = list(string)
  description = "CIDR blocks for public subnets."

  validation {
    condition     = length(var.public_cidr_blocks) == length(var.availability_zones)
    error_message = "public_cidr_blocks must contain one CIDR block per Availability Zone."
  }
}

variable "app_private_cidr_blocks" {
  type        = list(string)
  description = "CIDR blocks for private application subnets."

  validation {
    condition     = length(var.app_private_cidr_blocks) == length(var.availability_zones)
    error_message = "app_private_cidr_blocks must contain one CIDR block per Availability Zone."
  }
}

variable "db_private_cidr_blocks" {
  type        = list(string)
  description = "CIDR blocks for isolated database subnets."

  validation {
    condition     = length(var.db_private_cidr_blocks) == length(var.availability_zones)
    error_message = "db_private_cidr_blocks must contain one CIDR block per Availability Zone."
  }
}

variable "tags" {
  type        = map(string)
  description = "Tags applied to all subnets."
  default     = {}
}
