resource "aws_route53_zone" "this" {
  name    = var.name
  comment = ""
  tags    = var.tags
}
