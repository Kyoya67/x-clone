/*********************************************************************
 * 外部公開ALBとAPIタスクへの転送
 *********************************************************************/
resource "aws_lb" "api" {
  name                       = var.name
  internal                   = false
  load_balancer_type         = "application"
  subnets                    = var.public_subnet_ids
  security_groups            = [var.security_group_id]
  drop_invalid_header_fields = true
  tags                       = var.tags
}

# FargateのタスクIPはECSサービスが登録・解除する。
resource "aws_lb_target_group" "api" {
  name        = "${var.name}-targets"
  vpc_id      = var.vpc_id
  target_type = "ip"
  protocol    = "HTTP"
  port        = 8080

  health_check {
    path                = "/health"
    protocol            = "HTTP"
    port                = "traffic-port"
    matcher             = "200"
    interval            = 30
    timeout             = 5
    healthy_threshold   = 2
    unhealthy_threshold = 3
  }

  tags = var.tags
}
