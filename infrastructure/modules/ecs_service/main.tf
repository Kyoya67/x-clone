resource "aws_ecs_service" "this" {
  name            = var.name
  cluster         = var.cluster_arn
  task_definition = var.task_definition_arn
  launch_type     = "FARGATE"
  desired_count   = 1

  health_check_grace_period_seconds = 60

  load_balancer {
    target_group_arn = var.target_group_arn
    container_name   = "api"
    container_port   = 8080
  }

  # 更新時は旧タスクを維持して新タスクを起動する（一時的に最大2タスク）。
  deployment_minimum_healthy_percent = 100
  deployment_maximum_percent         = 200
  wait_for_steady_state              = true

  deployment_circuit_breaker {
    enable   = true
    rollback = true
  }

  network_configuration {
    subnets          = var.subnet_ids
    security_groups  = [var.security_group_id]
    assign_public_ip = false
  }

  lifecycle {
    # ECSサービスが参照するtask definition revisionはGitHub Actions CDが更新する。
    # Terraform applyで古いrevisionへ戻さないため、差分は無視する。
    ignore_changes = [task_definition]
  }

  tags = var.tags
}
