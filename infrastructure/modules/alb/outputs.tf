output "target_group_arn" {
  value = aws_lb_target_group.api.arn

  # ECSサービスの更新前に、ターゲットグループをALBへ関連付ける。
  depends_on = [aws_lb_listener.https]
}

output "load_balancer_arn" {
  value = aws_lb.api.arn
}

output "dns_name" {
  value = aws_lb.api.dns_name
}

output "zone_id" {
  value = aws_lb.api.zone_id
}
