output "arn" {
  value = aws_ecs_task_definition.this.arn
}

output "migration_arn" { value = aws_ecs_task_definition.migration.arn }
