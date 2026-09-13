module "vpc" {
  source     = "../modules/vpc"
  cidr_block = "10.0.0.0/16"
  tags       = local.common_tags
}

module "subnet" {
  source = "../modules/subnet"

  vpc_id             = module.vpc.id
  availability_zones = ["ap-northeast-1a", "ap-northeast-1c"]

  public_cidr_blocks  = ["10.0.0.0/18", "10.0.64.0/18"]
  private_cidr_blocks = ["10.0.128.0/18", "10.0.192.0/18"]

  tags = local.common_tags
}

module "internet_gateway" {
  source = "../modules/internet_gateway"

  vpc_id = module.vpc.id
  tags   = local.common_tags
}

module "route_table" {
  source = "../modules/route_table"

  vpc_id                   = module.vpc.id
  internet_gateway_id      = module.internet_gateway.id
  nat_network_interface_id = module.ec2.nat_network_interface_id
  public_subnet_ids        = module.subnet.public_ids
  private_subnet_ids       = module.subnet.private_ids
  tags                     = local.common_tags
}

module "route53" {
  source = "../modules/route53"

  name                                      = "stg.x-clone.kyo8.dev"
  api_certificate_domain_validation_options = module.api_certificate.domain_validation_options
  api_alias_record = {
    name     = "api-v1.stg.x-clone.kyo8.dev"
    dns_name = module.alb.dns_name
    zone_id  = module.alb.zone_id
  }
  amplify_certificate_dns_record = module.amplify.certificate_dns_record
  amplify_domain_dns_record      = module.amplify.domain_dns_record
  tags                           = local.common_tags
}

module "security_group" {
  source = "../modules/security_group"

  vpc_id              = module.vpc.id
  private_cidr_blocks = module.subnet.private_cidr_blocks
  tags                = local.common_tags
}

module "alb" {
  source = "../modules/alb"

  name              = "api"
  vpc_id            = module.vpc.id
  public_subnet_ids = module.subnet.public_ids
  security_group_id = module.security_group.api_alb_id
  domain_name       = "api-v1.stg.x-clone.kyo8.dev"
  certificate_arn   = module.api_certificate.validated_certificate_arn
  tags              = local.common_tags

  depends_on = [module.route_table]
}

module "api_certificate" {
  source = "../modules/acm_certificate"

  domain_name             = "api-v1.stg.x-clone.kyo8.dev"
  validation_record_fqdns = module.route53.api_certificate_validation_record_fqdns
  tags                    = local.common_tags
}

module "secrets_manager" {
  source = "../modules/secrets_manager"

  dbadmin = {
    name = "db/dbadmin"
  }
  dbadmin_password         = var.dbadmin_password
  dbadmin_password_version = var.dbadmin_password_version

  app_user = {
    name = "db/app_user"
  }
  migration_user = {
    name = "db/migration_user"
  }
  auth = {
    name = "auth/session"
  }
  auth_client_secret  = module.cognito.client_secret
  auth_session_secret = var.auth_session_secret
  auth_secret_version = var.auth_secret_version
  tags                = local.common_tags
}

module "cognito" {
  source = "../modules/cognito"

  name                 = "x-clone-stg"
  domain_prefix        = "x-clone-stg"
  callback_urls        = ["https://stg.x-clone.kyo8.dev/auth/callback"]
  logout_urls          = ["https://stg.x-clone.kyo8.dev/"]
  google_client_id     = var.google_client_id
  google_client_secret = var.google_client_secret
  tags                 = local.common_tags
}

module "iam" {
  source = "../modules/iam"

  repository_arn           = module.ecr.api_arn
  log_group_arn            = module.cloudwatch_logs.arn
  database_secret_arn      = module.secrets_manager.app_user_secret_arn
  auth_secret_arn          = module.secrets_manager.auth_secret_arn
  migration_repository_arn = module.ecr.db_migrator_arn
  migration_log_group_arn  = module.migration_logs.arn
  migration_secret_arn     = module.secrets_manager.migration_user_secret_arn
  github_repository        = "Kaminashi-Inc/ENG-1103_Kyoya67"
  github_oidc_subjects     = ["repo:Kaminashi-Inc@103104041/ENG-1103_Kyoya67@1356781382:ref:refs/heads/develop"]
  amplify_app_arn          = module.amplify.app_arn
  tags                     = local.common_tags
}

module "amplify" {
  source = "../modules/amplify"

  name        = "x-clone"
  branch_name = "stg"
  api_url     = "https://api-v1.stg.x-clone.kyo8.dev"
  domain_name = "stg.x-clone.kyo8.dev"
  tags        = local.common_tags
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
  name   = "x-clone"
  tags   = local.common_tags
}

module "ecs_task_definition" {
  source = "../modules/ecs_task_definition"

  backend = {
    family              = "api"
    image               = "${module.ecr.api_repository_url}:5045b0c-8-1"
    execution_role_arn  = module.iam.api_execution_role_arn
    task_role_arn       = module.iam.api_task_role_arn
    database_host       = module.rds.address
    database_secret_arn = module.secrets_manager.app_user_secret_arn
    auth_secret_arn     = module.secrets_manager.auth_secret_arn
    auth = {
      issuer         = module.cognito.issuer
      authorize_url  = module.cognito.authorize_url
      token_url      = module.cognito.token_url
      client_id      = module.cognito.client_id
      redirect_url   = "https://stg.x-clone.kyo8.dev/auth/callback"
      post_login_url = "https://stg.x-clone.kyo8.dev/"
      cookie_domain  = "stg.x-clone.kyo8.dev"
      cookie_secure  = "true"
    }
    log_group_name = module.cloudwatch_logs.name
  }

  migration = {
    family               = "db-migrator"
    image                = "${module.ecr.db_migrator_repository_url}:5045b0c-8-1"
    execution_role_arn   = module.iam.db_migrator_execution_role_arn
    task_role_arn        = module.iam.db_migrator_task_role_arn
    database_host        = module.rds.address
    migration_secret_arn = module.secrets_manager.migration_user_secret_arn
    log_group_name       = module.migration_logs.name
  }

  region = local.region
  tags   = local.common_tags
}

module "ecs_service" {
  source = "../modules/ecs_service"

  name                = "api"
  cluster_arn         = module.ecs.arn
  task_definition_arn = module.ecs_task_definition.api_arn
  target_group_arn    = module.alb.target_group_arn
  subnet_ids          = module.subnet.private_ids
  security_group_id   = module.security_group.api_id
  tags                = local.common_tags

  depends_on = [module.iam, module.route_table, module.security_group]
}

module "rds" {
  source = "../modules/rds"

  depends_on = [module.secrets_manager]

  dbadmin_password         = var.dbadmin_password
  dbadmin_password_version = var.dbadmin_password_version

  identifier        = "app-db"
  engine_version    = "16.15"
  instance_class    = "db.t4g.micro"
  multi_az          = false
  subnet_ids        = module.subnet.private_ids
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
