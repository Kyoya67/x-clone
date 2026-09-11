variable "image_count" {
  type        = number
  description = "Number of recent images retained by the lifecycle policy."
  default     = 3

  validation {
    condition     = var.image_count >= 1 && floor(var.image_count) == var.image_count
    error_message = "image_count must be a positive integer."
  }
}

variable "tags" {
  type        = map(string)
  description = "Tags applied to the repository."
  default     = {}
}
