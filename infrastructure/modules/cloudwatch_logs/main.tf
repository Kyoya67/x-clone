resource "aws_cloudwatch_log_group" "this" {
  name              = var.name
  retention_in_days = 14
  tags              = var.tags
}
