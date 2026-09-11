output "arn" {
  value = aws_ecs_task_definition.api.arn
}

output "migration_arn" { value = aws_ecs_task_definition.migration.arn }
