data "aws_iam_policy_document" "ecs_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_openid_connect_provider" "github_actions" {
  url            = "https://token.actions.githubusercontent.com"
  client_id_list = ["sts.amazonaws.com"]
  tags           = var.tags
}

data "aws_iam_policy_document" "github_actions_assume_role" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [aws_iam_openid_connect_provider.github_actions.arn]
    }

    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:aud"
      values   = ["sts.amazonaws.com"]
    }

    condition {
      test     = "StringLike"
      variable = "token.actions.githubusercontent.com:sub"
      values   = var.github_oidc_subjects
    }
  }
}

/*********************************************************************
 * API用IAMロール・ポリシー
 *********************************************************************/
resource "aws_iam_role" "api_execution" {
  lifecycle {
    create_before_destroy = true
  }

  name               = "api-task-execution"
  assume_role_policy = data.aws_iam_policy_document.ecs_assume_role.json
  tags               = var.tags
}

# ECS基盤が起動時に使用する権限。アプリ本体には付与しない。
resource "aws_iam_policy" "api_execution" {
  lifecycle {
    create_before_destroy = true
  }

  name = "api-task-execution"
  tags = var.tags
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["ecr:GetAuthorizationToken"]
        Resource = "*"
      },
      {
        Effect   = "Allow"
        Action   = ["ecr:BatchCheckLayerAvailability", "ecr:GetDownloadUrlForLayer", "ecr:BatchGetImage"]
        Resource = var.repository_arn
      },
      {
        Effect   = "Allow"
        Action   = ["logs:CreateLogStream", "logs:PutLogEvents"]
        Resource = "${var.log_group_arn}:*"
      },
      {
        Effect   = "Allow"
        Action   = ["secretsmanager:GetSecretValue"]
        Resource = var.database_secret_arn
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "api_execution" {
  lifecycle {
    create_before_destroy = true
  }

  role       = aws_iam_role.api_execution.name
  policy_arn = aws_iam_policy.api_execution.arn
}

# Goアプリは現在AWS APIを呼び出さないため、権限ポリシーを付けない。
resource "aws_iam_role" "api_task" {
  lifecycle {
    create_before_destroy = true
  }

  name               = "api-task"
  assume_role_policy = data.aws_iam_policy_document.ecs_assume_role.json
  tags               = var.tags
}

/*********************************************************************
 * マイグレーション用IAMロール・ポリシー
 *********************************************************************/
# 管理者Secretへの権限は通常のバックエンド実行ロールに付けない。
resource "aws_iam_role" "db_migrator_execution" {
  lifecycle {
    create_before_destroy = true
  }

  name               = "db-migrator-task-execution"
  assume_role_policy = data.aws_iam_policy_document.ecs_assume_role.json
  tags               = var.tags
}

resource "aws_iam_policy" "db_migrator_execution" {
  lifecycle {
    create_before_destroy = true
  }

  name = "db-migrator-task-execution"
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
      { Effect = "Allow", Action = ["secretsmanager:GetSecretValue"], Resource = var.migration_secret_arn }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "db_migrator_execution" {
  lifecycle {
    create_before_destroy = true
  }

  role       = aws_iam_role.db_migrator_execution.name
  policy_arn = aws_iam_policy.db_migrator_execution.arn
}

# コンテナ本体はAWS APIを呼ばない。DBの権限はmigration_userのSQL権限。
resource "aws_iam_role" "db_migrator_task" {
  lifecycle {
    create_before_destroy = true
  }

  name               = "db-migrator-task"
  assume_role_policy = data.aws_iam_policy_document.ecs_assume_role.json
  tags               = var.tags
}

/*********************************************************************
 * NATインスタンス用IAMロール・インスタンスプロファイル
 *********************************************************************/
resource "aws_iam_role" "nat_ssm" {
  name = "nat-ssm"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Action    = "sts:AssumeRole"
      Principal = { Service = "ec2.amazonaws.com" }
    }]
  })
  tags = var.tags
}

# SSM Agent用のAWS管理ポリシー。DBやSecrets Managerへの権限は付与しない。
resource "aws_iam_role_policy_attachment" "nat_ssm" {
  role       = aws_iam_role.nat_ssm.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_instance_profile" "nat_ssm" {
  name = "nat-ssm"
  role = aws_iam_role.nat_ssm.name
  tags = var.tags
}

/*********************************************************************
 * GitHub Actions CD用IAMロール・ポリシー
 *********************************************************************/
resource "aws_iam_role" "github_actions_cd" {
  name               = "github-actions-cd"
  assume_role_policy = data.aws_iam_policy_document.github_actions_assume_role.json
  tags               = var.tags
}

resource "aws_iam_policy" "github_actions_cd" {
  name = "github-actions-cd"
  tags = var.tags
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "amplify:CreateDeployment",
          "amplify:StartDeployment",
          "amplify:GetJob"
        ]
        Resource = [
          var.amplify_app_arn,
          "${var.amplify_app_arn}/branches/*",
          "${var.amplify_app_arn}/jobs/*"
        ]
      },
      {
        Effect   = "Allow"
        Action   = ["ecr:GetAuthorizationToken"]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "ecr:BatchCheckLayerAvailability",
          "ecr:CompleteLayerUpload",
          "ecr:InitiateLayerUpload",
          "ecr:PutImage",
          "ecr:UploadLayerPart"
        ]
        Resource = [
          var.repository_arn,
          var.migration_repository_arn
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "ecs:DescribeTaskDefinition",
          "ecs:RegisterTaskDefinition",
          "ecs:RunTask",
          "ecs:DescribeTasks",
          "ecs:UpdateService",
          "ecs:DescribeServices"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "ec2:DescribeSubnets",
          "ec2:DescribeSecurityGroups"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = ["iam:PassRole"]
        Resource = [
          aws_iam_role.api_execution.arn,
          aws_iam_role.api_task.arn,
          aws_iam_role.db_migrator_execution.arn,
          aws_iam_role.db_migrator_task.arn
        ]
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "github_actions_cd" {
  role       = aws_iam_role.github_actions_cd.name
  policy_arn = aws_iam_policy.github_actions_cd.arn
}
