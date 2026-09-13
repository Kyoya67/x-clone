# 通常バックエンド。マイグレーションとは別リソースとして管理する。
resource "aws_ecs_task_definition" "api" {
  family                   = var.backend.family
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = var.backend.execution_role_arn
  task_role_arn            = var.backend.task_role_arn

  runtime_platform {
    operating_system_family = "LINUX"
    cpu_architecture        = "X86_64"
  }

  container_definitions = jsonencode([{
    name                   = "api"
    image                  = var.backend.image
    essential              = true
    user                   = "65532:65532"
    readonlyRootFilesystem = true
    portMappings           = [{ containerPort = 8080, protocol = "tcp" }]
    # scratchにはシェル・curlがないため、Go製コマンドを直接実行する。
    healthCheck = {
      command     = ["CMD", "/app/healthcheck"]
      interval    = 30
      timeout     = 5
      retries     = 3
      startPeriod = 30
    }
    environment = [
      { name = "PORT", value = "8080" },
      { name = "DB_HOST", value = var.backend.database_host },
      { name = "DB_PORT", value = "5432" },
      { name = "AUTH_ISSUER", value = var.backend.auth.issuer },
      { name = "AUTH_AUTHORIZE_URL", value = var.backend.auth.authorize_url },
      { name = "AUTH_TOKEN_URL", value = var.backend.auth.token_url },
      { name = "AUTH_CLIENT_ID", value = var.backend.auth.client_id },
      { name = "AUTH_REDIRECT_URL", value = var.backend.auth.redirect_url },
      { name = "AUTH_POST_LOGIN_URL", value = var.backend.auth.post_login_url },
      { name = "AUTH_COOKIE_DOMAIN", value = var.backend.auth.cookie_domain },
      { name = "AUTH_COOKIE_SECURE", value = var.backend.auth.cookie_secure },
      { name = "AUTH_COOKIE_NAME_PREFIX", value = var.backend.auth.cookie_prefix }
    ]
    secrets = [
      { name = "DB_USER", valueFrom = "${var.backend.database_secret_arn}:username::" },
      { name = "DB_PASSWORD", valueFrom = "${var.backend.database_secret_arn}:password::" },
      { name = "AUTH_CLIENT_SECRET", valueFrom = "${var.backend.auth_secret_arn}:client_secret::" },
      { name = "AUTH_SESSION_SECRET", valueFrom = "${var.backend.auth_secret_arn}:session_secret::" }
    ]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        awslogs-group         = var.backend.log_group_name
        awslogs-region        = var.region
        awslogs-stream-prefix = "backend"
      }
    }
  }])

  tags = var.tags
}

# サービスとして常駐させず、RunTaskで1回だけ実行する。
resource "aws_ecs_task_definition" "migration" {
  family                   = var.migration.family
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = var.migration.execution_role_arn
  task_role_arn            = var.migration.task_role_arn
  runtime_platform {
    operating_system_family = "LINUX"
    cpu_architecture        = "X86_64"
  }
  container_definitions = jsonencode([{
    name                   = "migration"
    image                  = var.migration.image
    essential              = true
    user                   = "65532:65532"
    readonlyRootFilesystem = true
    # デフォルトは状態確認。適用時は起動コマンドがupで上書きする。
    command = ["--action", "status"]
    environment = [
      { name = "DB_HOST", value = var.migration.database_host },
      { name = "DB_PORT", value = "5432" }
    ]
    secrets = [
      { name = "DB_USER", valueFrom = "${var.migration.migration_secret_arn}:username::" },
      { name = "DB_PASSWORD", valueFrom = "${var.migration.migration_secret_arn}:password::" }
    ]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        awslogs-group         = var.migration.log_group_name
        awslogs-region        = var.region
        awslogs-stream-prefix = "migration"
      }
    }
  }])

  tags = var.tags
}
