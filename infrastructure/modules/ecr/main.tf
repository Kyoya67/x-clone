module "backend" {
  source      = "../ecr_unit"
  name        = "backend"
  image_count = var.image_count
  tags        = var.tags
}
module "migration" {
  source      = "../ecr_unit"
  name        = "backend-migration"
  image_count = var.image_count
  tags        = var.tags
}
