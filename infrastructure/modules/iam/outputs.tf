output "api_execution_role_arn" {
  value      = aws_iam_role.api_execution.arn
  depends_on = [aws_iam_role_policy_attachment.api_execution]
}
output "api_task_role_arn" {
  value = aws_iam_role.api_task.arn
}

output "api_execution_policy_arn" {
  value       = aws_iam_policy.api_execution.arn
  description = "Customer-managed policy ARN available for attachment to other roles."
}
output "nat_instance_profile_name" {
  value      = aws_iam_instance_profile.nat_ssm.name
  depends_on = [aws_iam_role_policy_attachment.nat_ssm]
}

output "db_migrator_execution_role_arn" {
  value      = aws_iam_role.db_migrator_execution.arn
  depends_on = [aws_iam_role_policy_attachment.db_migrator_execution]
}
output "db_migrator_task_role_arn" { value = aws_iam_role.db_migrator_task.arn }
