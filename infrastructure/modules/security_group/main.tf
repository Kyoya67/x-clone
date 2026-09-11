data "aws_region" "current" {}

# APIタスクに割り当てるSG。現時点ではNATホストからの確認のみ許可する。
resource "aws_security_group" "backend" {
  name        = "backend"
  description = "Backend ECS tasks"
  vpc_id      = var.vpc_id
  tags        = merge(var.tags, { Name = "backend-sg" })
}

resource "aws_security_group" "db" {
  name        = "db"
  description = "PostgreSQL access from backend tasks only"
  vpc_id      = var.vpc_id
  tags        = merge(var.tags, { Name = "db-sg" })
}

# SSMで接続したNATホストからAPIのhealthを確認する。外部公開はしない。
resource "aws_vpc_security_group_ingress_rule" "backend_from_nat" {
  security_group_id            = aws_security_group.backend.id
  referenced_security_group_id = aws_security_group.nat.id
  ip_protocol                  = "tcp"
  from_port                    = 8080
  to_port                      = 8080
}

# SSM転送はNATホスト自身からRDSへの新しい接続になる。
resource "aws_vpc_security_group_ingress_rule" "db_from_nat" {
  security_group_id            = aws_security_group.db.id
  referenced_security_group_id = aws_security_group.nat.id
  ip_protocol                  = "tcp"
  from_port                    = 5432
  to_port                      = 5432
}

resource "aws_vpc_security_group_ingress_rule" "db_from_backend" {
  security_group_id            = aws_security_group.db.id
  referenced_security_group_id = aws_security_group.backend.id
  ip_protocol                  = "tcp"
  from_port                    = 5432
  to_port                      = 5432
}

resource "aws_vpc_security_group_egress_rule" "backend_to_db" {
  security_group_id            = aws_security_group.backend.id
  referenced_security_group_id = aws_security_group.db.id
  ip_protocol                  = "tcp"
  from_port                    = 5432
  to_port                      = 5432
}

# イメージ取得・ログ送信・Secrets Managerへのアクセスに使用する。
resource "aws_vpc_security_group_egress_rule" "backend_https" {
  security_group_id = aws_security_group.backend.id
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "tcp"
  from_port         = 443
  to_port           = 443
}

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
  for_each = toset(var.private_cidr_blocks)

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
