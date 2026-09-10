data "aws_region" "current" {}

data "aws_ec2_managed_prefix_list" "instance_connect" {
  name = "com.amazonaws.${data.aws_region.current.region}.ec2-instance-connect"
}

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

resource "aws_vpc_security_group_ingress_rule" "nat_instance_connect" {
  security_group_id = aws_security_group.nat.id
  description       = "SSH from the EC2 Instance Connect service"
  prefix_list_id    = data.aws_ec2_managed_prefix_list.instance_connect.id
  ip_protocol       = "tcp"
  from_port         = 22
  to_port           = 22
}

resource "aws_vpc_security_group_egress_rule" "nat" {
  security_group_id = aws_security_group.nat.id
  description       = "Internet-bound traffic from private application workloads"
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "-1"
}
