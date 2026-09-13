resource "aws_route53_zone" "this" {
  name    = var.name
  comment = ""
  tags    = var.tags
}

resource "aws_route53_record" "api_certificate_validation" {
  for_each = {
    for option in var.api_certificate_domain_validation_options : option.domain_name => option
  }

  zone_id = aws_route53_zone.this.zone_id
  name    = each.value.resource_record_name
  type    = each.value.resource_record_type
  records = [each.value.resource_record_value]
  ttl     = 300
}

resource "aws_route53_record" "api" {
  count = var.api_alias_record == null ? 0 : 1

  zone_id = aws_route53_zone.this.zone_id
  name    = var.api_alias_record.name
  type    = "A"

  alias {
    name                   = var.api_alias_record.dns_name
    zone_id                = var.api_alias_record.zone_id
    evaluate_target_health = true
  }
}

resource "aws_route53_record" "amplify_certificate_validation" {
  count = var.amplify_certificate_dns_record == null ? 0 : 1

  zone_id = aws_route53_zone.this.zone_id
  name    = var.amplify_certificate_dns_record.name
  type    = var.amplify_certificate_dns_record.type
  records = [var.amplify_certificate_dns_record.value]
  ttl     = 300

  # Amplifyが同じ検証レコードを自動作成した場合も、Terraform管理へ引き継ぐ。
  allow_overwrite = true
}

resource "aws_route53_record" "amplify_frontend" {
  count = var.amplify_domain_dns_record == null ? 0 : 1

  zone_id = aws_route53_zone.this.zone_id
  name    = var.amplify_domain_dns_record.name
  type    = "A"

  # Amplify側が同じ配信用レコードを自動作成する場合にも対応する。
  allow_overwrite = true

  alias {
    name                   = var.amplify_domain_dns_record.value
    zone_id                = "Z2FDTNDATAQYW2" # CloudFront共通のホストゾーンID
    evaluate_target_health = false
  }
}
