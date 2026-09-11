output "execution_role_arn" {
  value      = aws_iam_role.execution.arn
  depends_on = [aws_iam_role_policy_attachment.execution]
}
output "task_role_arn" {
  value = aws_iam_role.task.arn
}

output "execution_policy_arn" {
  value       = aws_iam_policy.execution.arn
  description = "Customer-managed policy ARN available for attachment to other roles."
}
output "nat_instance_profile_name" {
  value      = aws_iam_instance_profile.nat_ssm.name
  depends_on = [aws_iam_role_policy_attachment.nat_ssm]
}

output "migration_execution_role_arn" {
  value      = aws_iam_role.migration_execution.arn
  depends_on = [aws_iam_role_policy_attachment.migration_execution]
}
output "migration_task_role_arn" { value = aws_iam_role.migration_task.arn }
