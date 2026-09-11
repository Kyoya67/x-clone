# 管理者Secretへの権限は通常のバックエンド実行ロールに付けない。
resource "aws_iam_role" "migration_execution" {
  name               = "migration-task-execution"
  assume_role_policy = data.aws_iam_policy_document.ecs_assume_role.json
  tags               = var.tags
}

resource "aws_iam_policy" "migration_execution" {
  name = "migration-task-execution"
  tags = var.tags
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      { Effect = "Allow", Action = ["ecr:GetAuthorizationToken"], Resource = "*" },
      {
        Effect   = "Allow"
        Action   = ["ecr:BatchCheckLayerAvailability", "ecr:GetDownloadUrlForLayer", "ecr:BatchGetImage"]
        Resource = var.migration_repository_arn
      },
      {
        Effect   = "Allow"
        Action   = ["logs:CreateLogStream", "logs:PutLogEvents"]
        Resource = "${var.migration_log_group_arn}:*"
      },
      { Effect = "Allow", Action = ["secretsmanager:GetSecretValue"], Resource = var.admin_secret_arn }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "migration_execution" {
  role       = aws_iam_role.migration_execution.name
  policy_arn = aws_iam_policy.migration_execution.arn
}

# コンテナ本体はAWS APIを呼ばない。DBの権限はdbadminのSQL権限。
resource "aws_iam_role" "migration_task" {
  name               = "migration-task"
  assume_role_policy = data.aws_iam_policy_document.ecs_assume_role.json
  tags               = var.tags
}
