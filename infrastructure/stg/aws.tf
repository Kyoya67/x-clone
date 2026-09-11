module "vpc" {
  source     = "../modules/vpc"
  cidr_block = "10.0.0.0/16"
  tags       = local.common_tags
}

module "subnet" {
  source = "../modules/subnet"

  vpc_id             = module.vpc.id
  availability_zones = ["ap-northeast-1a", "ap-northeast-1c"]

  public_cidr_blocks      = ["10.0.0.0/24", "10.0.1.0/24"]
  app_private_cidr_blocks = ["10.0.10.0/24", "10.0.11.0/24"]
  db_private_cidr_blocks  = ["10.0.20.0/24", "10.0.21.0/24"]

  tags = local.common_tags
}

module "route_table" {
  source = "../modules/route_table"

  vpc_id                   = module.vpc.id
  internet_gateway_id      = module.internet_gateway.id
  nat_network_interface_id = module.ec2.nat_network_interface_id
  public_subnet_ids        = module.subnet.public_ids
  app_private_subnet_ids   = module.subnet.app_private_ids
  db_private_subnet_ids    = module.subnet.db_private_ids
  tags                     = local.common_tags
}

module "internet_gateway" {
  source = "../modules/internet_gateway"

  vpc_id = module.vpc.id
  tags   = local.common_tags
}

module "security_group" {
  source = "../modules/security_group"

  vpc_id                  = module.vpc.id
  app_private_cidr_blocks = module.subnet.app_private_cidr_blocks
  tags                    = local.common_tags
}

module "iam" {
  source = "../modules/iam"

  repository_arn           = module.ecr.backend_arn
  log_group_arn            = module.cloudwatch_logs.arn
  database_url_secret_arn  = module.secrets_manager.arn
  migration_repository_arn = module.ecr.migration_arn
  migration_log_group_arn  = module.migration_logs.arn
  admin_secret_arn         = module.rds.master_user_secret_arn
  tags                     = local.common_tags
}

module "secrets_manager" {
  source = "../modules/secrets_manager"

  name = "backend/database-url"
  tags = local.common_tags
}

module "ec2" {
  source = "../modules/ec2"

  public_subnet_id      = module.subnet.public_ids[0]
  nat_security_group_id = module.security_group.nat_id
  instance_profile_name = module.iam.nat_instance_profile_name
  tags                  = local.common_tags
}

module "ecr" {
  source      = "../modules/ecr"
  image_count = 3
  tags        = local.common_tags
}

module "ecs" {
  source = "../modules/ecs"
  name   = "app"
  tags   = local.common_tags
}

module "ecs_task_definition" {
  source = "../modules/ecs_task_definition"

  backend = {
    family                  = "backend"
    image                   = "${module.ecr.backend_repository_url}:78694d"
    execution_role_arn      = module.iam.execution_role_arn
    task_role_arn           = module.iam.task_role_arn
    database_url_secret_arn = module.secrets_manager.arn
    log_group_name          = module.cloudwatch_logs.name
  }

  migration = {
    family             = "backend-migration"
    image              = "${module.ecr.migration_repository_url}:${var.migration_image_tag}"
    execution_role_arn = module.iam.migration_execution_role_arn
    task_role_arn      = module.iam.migration_task_role_arn
    database_host      = module.rds.address
    admin_secret_arn   = module.rds.master_user_secret_arn
    log_group_name     = module.migration_logs.name
  }

  region = local.region
  tags   = local.common_tags
}

module "rds" {
  source = "../modules/rds"

  identifier        = "app-db"
  engine_version    = "16.15"
  instance_class    = "db.t4g.micro"
  multi_az          = false
  subnet_ids        = module.subnet.db_private_ids
  security_group_id = module.security_group.db_id
  tags              = local.common_tags
}

module "migration_logs" {
  source = "../modules/cloudwatch_logs"
  name   = "/ecs/backend-migration"
  tags   = local.common_tags
}

module "cloudwatch_logs" {
  source = "../modules/cloudwatch_logs"

  name = "/ecs/backend"
  tags = local.common_tags
}
