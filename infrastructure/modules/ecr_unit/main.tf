resource "aws_ecr_repository" "this" {
  name                 = var.name
  image_tag_mutability = "IMMUTABLE"
  force_delete         = false

  tags = var.tags
}

resource "aws_ecr_lifecycle_policy" "this" {
  repository = aws_ecr_repository.this.name

  # タグの有無を問わず、push日時が新しい順に指定件数を保持する。
  # ECSで使用中かどうかは判定されない。
  policy = jsonencode({
    rules = [{
      rulePriority = 1
      description  = "Keep the latest ${var.image_count} images"
      selection = {
        tagStatus   = "any"
        countType   = "imageCountMoreThan"
        countNumber = var.image_count
      }
      action = {
        type = "expire"
      }
    }]
  })
}
