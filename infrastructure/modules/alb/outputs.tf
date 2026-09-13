output "target_group_arn" {
  value = aws_lb_target_group.api.arn

  # ECSサービスの更新前に、ターゲットグループをALBへ関連付ける。
  depends_on = [aws_lb_listener.https]
}

output "api_url" {
  value = "https://${var.domain_name}"
}
