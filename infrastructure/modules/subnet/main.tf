locals {
  subnet_by_az = {
    for index, availability_zone in var.availability_zones : availability_zone => {
      public_cidr      = var.public_cidr_blocks[index]
      app_private_cidr = var.app_private_cidr_blocks[index]
      db_private_cidr  = var.db_private_cidr_blocks[index]
    }
  }
}

resource "aws_subnet" "public" {
  for_each = local.subnet_by_az

  vpc_id                  = var.vpc_id
  availability_zone       = each.key
  cidr_block              = each.value.public_cidr
  map_public_ip_on_launch = true

  tags = merge(var.tags, {
    Name = "public-${each.key}"
    Tier = "public"
  })
}

resource "aws_subnet" "app_private" {
  for_each = local.subnet_by_az

  vpc_id            = var.vpc_id
  availability_zone = each.key
  cidr_block        = each.value.app_private_cidr

  tags = merge(var.tags, {
    Name = "app-private-${each.key}"
    Tier = "app-private"
  })
}

resource "aws_subnet" "db_private" {
  for_each = local.subnet_by_az

  vpc_id            = var.vpc_id
  availability_zone = each.key
  cidr_block        = each.value.db_private_cidr

  tags = merge(var.tags, {
    Name = "db-private-${each.key}"
    Tier = "db-private"
  })
}
