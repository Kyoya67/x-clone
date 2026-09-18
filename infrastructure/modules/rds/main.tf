resource "aws_db_subnet_group" "this" {
  name       = var.identifier
  subnet_ids = var.subnet_ids
  tags       = merge(var.tags, { Name = "db-subnet-group" })
}

resource "aws_db_parameter_group" "this" {
  name   = "${var.identifier}-postgres16"
  family = "postgres16"

  # 接続側のverify-fullとは別に、DB側でも平文接続を拒否する。
  parameter {
    name         = "rds.force_ssl"
    value        = "1"
    apply_method = "pending-reboot"
  }

  tags = var.tags
}

resource "aws_db_instance" "this" {
  identifier     = var.identifier
  engine         = "postgres"
  engine_version = var.engine_version
  instance_class = var.instance_class
  db_name        = "app"
  username       = "dbadmin"
  port           = 5432

  # .envで指定したパスワードを使用する。Stateには保存しない。
  password_wo         = var.dbadmin_password
  password_wo_version = var.dbadmin_password_version

  allocated_storage     = 20
  max_allocated_storage = 100
  storage_type          = "gp3"
  storage_encrypted     = true

  db_subnet_group_name   = aws_db_subnet_group.this.name
  parameter_group_name   = aws_db_parameter_group.this.name
  vpc_security_group_ids = [var.security_group_id]
  publicly_accessible    = false
  multi_az               = var.multi_az
  ca_cert_identifier     = "rds-ca-rsa2048-g1"

  backup_retention_period    = 7
  backup_window              = "18:00-19:00"
  maintenance_window         = "sun:19:00-sun:20:00"
  auto_minor_version_upgrade = true
  copy_tags_to_snapshot      = true
  deletion_protection        = false
  skip_final_snapshot        = false
  final_snapshot_identifier  = "${var.identifier}-final"

  tags = merge(var.tags, { Name = var.identifier })
}
