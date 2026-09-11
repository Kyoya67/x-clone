# モジュール構成変更のみ。既存リソースを再作成せずState上のアドレスを引き継ぐ。
moved {
  from = module.ecs_task_definition.aws_ecs_task_definition.this
  to   = module.ecs_task_definition.aws_ecs_task_definition.api
}

moved {
  from = module.ecr.aws_ecr_repository.this
  to   = module.ecr.module.backend.aws_ecr_repository.this
}
moved {
  from = module.ecr.aws_ecr_lifecycle_policy.this
  to   = module.ecr.module.backend.aws_ecr_lifecycle_policy.this
}
moved {
  from = module.migration_ecr
  to   = module.ecr.module.migration
}
moved {
  from = module.ecs_cluster
  to   = module.ecs
}
moved {
  from = module.ecs_migration_task_definition.aws_ecs_task_definition.this
  to   = module.ecs_task_definition.aws_ecs_task_definition.migration
}
