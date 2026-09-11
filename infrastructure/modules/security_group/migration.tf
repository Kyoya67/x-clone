resource "aws_security_group" "migration" {
  name        = "migration"
  description = "Standalone migration task; no inbound access"
  vpc_id      = var.vpc_id
  tags        = merge(var.tags, { Name = "migration-sg" })
}

resource "aws_vpc_security_group_egress_rule" "migration_to_db" {
  security_group_id            = aws_security_group.migration.id
  referenced_security_group_id = aws_security_group.db.id
  ip_protocol                  = "tcp"
  from_port                    = 5432
  to_port                      = 5432
}

resource "aws_vpc_security_group_ingress_rule" "db_from_migration" {
  security_group_id            = aws_security_group.db.id
  referenced_security_group_id = aws_security_group.migration.id
  ip_protocol                  = "tcp"
  from_port                    = 5432
  to_port                      = 5432
}

resource "aws_vpc_security_group_egress_rule" "migration_https" {
  security_group_id = aws_security_group.migration.id
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "tcp"
  from_port         = 443
  to_port           = 443
}
