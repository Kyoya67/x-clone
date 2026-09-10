resource "aws_security_group" "nat" {
  name        = "nat-instance"
  description = "Allow outbound traffic from private application subnets through the NAT instance"
  vpc_id      = var.vpc_id

  lifecycle {
    create_before_destroy = true
  }

  tags = merge(var.tags, {
    Name = "nat-instance-sg"
  })
}

resource "aws_vpc_security_group_ingress_rule" "nat" {
  for_each = toset(var.app_private_cidr_blocks)

  security_group_id = aws_security_group.nat.id
  description       = "Traffic initiated by private application workloads"
  cidr_ipv4         = each.value
  ip_protocol       = "-1"
}

resource "aws_vpc_security_group_egress_rule" "nat" {
  security_group_id = aws_security_group.nat.id
  description       = "Internet-bound traffic from private application workloads"
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "-1"
}
