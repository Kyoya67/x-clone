output "target_group_arn" {
  value = aws_lb_target_group.api.arn
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
