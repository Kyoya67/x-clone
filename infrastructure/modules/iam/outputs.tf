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
