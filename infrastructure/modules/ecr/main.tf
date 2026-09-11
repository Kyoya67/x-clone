module "api" {
  source      = "../ecr_unit"
  name        = "api"
  image_count = var.image_count
  tags        = var.tags
}

module "db_migrator" {
  source      = "../ecr_unit"
  name        = "db-migrator"
  image_count = var.image_count
  tags        = var.tags
}
