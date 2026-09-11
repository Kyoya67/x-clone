locals {
  subnet_by_az = {
    for index, availability_zone in var.availability_zones : availability_zone => {
      public_cidr  = var.public_cidr_blocks[index]
      private_cidr = var.private_cidr_blocks[index]
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
    Name = "public-${replace(each.key, "ap-northeast-", "")}"
    Tier = "public"
  })
}

resource "aws_subnet" "private" {
  for_each = local.subnet_by_az

  vpc_id            = var.vpc_id
  availability_zone = each.key
  cidr_block        = each.value.private_cidr

  tags = merge(var.tags, {
    Name = "private-${replace(each.key, "ap-northeast-", "")}"
    Tier = "private"
  })
}
