# RDSが使用中のサブネット・関連付けを引き継ぎ、再作成を避ける。
moved {
  from = module.subnet.aws_subnet.db_private
  to   = module.subnet.aws_subnet.private
}

moved {
  from = module.route_table.aws_route_table.db_private
  to   = module.route_table.aws_route_table.private
}

moved {
  from = module.route_table.aws_route_table_association.db_private
  to   = module.route_table.aws_route_table_association.private
}
