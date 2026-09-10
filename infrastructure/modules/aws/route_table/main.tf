resource "aws_route_table" "public" {
  vpc_id = var.vpc_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = var.internet_gateway_id
  }

  tags = merge(var.tags, {
    Name = "public-rt"
  })
}

resource "aws_route_table_association" "public" {
  count = length(var.public_subnet_ids)

  subnet_id      = var.public_subnet_ids[count.index]
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table" "app_private" {
  vpc_id = var.vpc_id

  route {
    cidr_block           = "0.0.0.0/0"
    network_interface_id = var.nat_network_interface_id
  }

  tags = merge(var.tags, {
    Name = "app-private-rt"
  })
}

resource "aws_route_table_association" "app_private" {
  count = length(var.app_private_subnet_ids)

  subnet_id      = var.app_private_subnet_ids[count.index]
  route_table_id = aws_route_table.app_private.id
}

resource "aws_route_table" "db_private" {
  vpc_id = var.vpc_id

  tags = merge(var.tags, {
    Name = "db-private-rt"
  })
}

resource "aws_route_table_association" "db_private" {
  count = length(var.db_private_subnet_ids)

  subnet_id      = var.db_private_subnet_ids[count.index]
  route_table_id = aws_route_table.db_private.id
}
