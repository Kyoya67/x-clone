output "repository_url" {
  value       = aws_ecr_repository.this.repository_url
  description = "Repository URL used for Docker push and ECS image references."
}

output "arn" {
  value       = aws_ecr_repository.this.arn
  description = "Repository ARN used in IAM policies."
}
