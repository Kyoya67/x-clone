module "vpc" {
  source = "../modules/aws/vpc"

  name       = "x-clone-${local.env}"
  cidr_block = "10.0.0.0/16"

  tags = local.common_tags
}

module "subnet" {
  source = "../modules/aws/subnet"

  name               = "x-clone-${local.env}"
  vpc_id             = module.vpc.id
  availability_zones = ["ap-northeast-1a", "ap-northeast-1c"]

  public_cidr_blocks      = ["10.0.0.0/24", "10.0.1.0/24"]
  app_private_cidr_blocks = ["10.0.10.0/24", "10.0.11.0/24"]
  db_private_cidr_blocks  = ["10.0.20.0/24", "10.0.21.0/24"]

  tags = local.common_tags
}

module "internet_gateway" {
  source = "../modules/aws/internet_gateway"

  name   = "x-clone-${local.env}"
  vpc_id = module.vpc.id
  tags   = local.common_tags
}

module "security_group" {
  source = "../modules/aws/security_group"

  name                    = "x-clone-${local.env}"
  vpc_id                  = module.vpc.id
  app_private_cidr_blocks = module.subnet.app_private_cidr_blocks
  tags                    = local.common_tags
}

module "ec2" {
  source = "../modules/aws/ec2"

  name                  = "x-clone-${local.env}"
  public_subnet_id      = module.subnet.public_ids[0]
  nat_security_group_id = module.security_group.nat_id
  tags                  = local.common_tags
}

module "route_table" {
  source = "../modules/aws/route_table"

  name                     = "x-clone-${local.env}"
  vpc_id                   = module.vpc.id
  internet_gateway_id      = module.internet_gateway.id
  nat_network_interface_id = module.ec2.nat_network_interface_id
  public_subnet_ids        = module.subnet.public_ids
  app_private_subnet_ids   = module.subnet.app_private_ids
  db_private_subnet_ids    = module.subnet.db_private_ids
  tags                     = local.common_tags
}
