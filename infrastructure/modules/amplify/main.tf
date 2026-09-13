resource "aws_amplify_app" "this" {
  name     = var.name
  platform = "WEB"

  enable_auto_branch_creation = false
  enable_branch_auto_build    = false

  # /api/posts → APIの/posts。ブラウザのURLを変えずにHTTPSで転送する。
  # SPAのルールより先に評価させ、APIへの応答をindex.htmlに置き換えない。
  custom_rule {
    source = "/api/<*>"
    target = "${var.api_url}/<*>"
    status = "200"
  }

  # 画面URLをSPAへ渡す。APIと拡張子付きの静的ファイルは対象外。
  custom_rule {
    source = "</^(?!/api(?:/|$))[^.]*$/>"
    target = "/index.html"
    status = "200"
  }

  tags = var.tags
}

resource "aws_amplify_branch" "this" {
  app_id                      = aws_amplify_app.this.id
  branch_name                 = var.branch_name
  framework                   = "React"
  stage                       = "DEVELOPMENT"
  enable_auto_build           = false
  enable_pull_request_preview = false
  tags                        = var.tags
}

# Git連携なしでも、デプロイ済みのstgブランチに独自ドメインを関連付けられる。
resource "aws_amplify_domain_association" "this" {
  app_id                 = aws_amplify_app.this.id
  domain_name            = var.domain_name
  enable_auto_sub_domain = false

  # この後のRoute 53レコード作成を先に進めるため、DNS検証完了をここでは待たない。
  wait_for_verification = false

  certificate_settings {
    type = "AMPLIFY_MANAGED"
  }

  sub_domain {
    branch_name = aws_amplify_branch.this.branch_name
    prefix      = ""
  }
}

locals {
  # Amplifyが返す「名前 CNAME 値」を分解する。
  certificate_dns_record = regexall("\\S+", aws_amplify_domain_association.this.certificate_verification_dns_record)
  # sub_domainは1件。「 CNAME 配信先」からCloudFrontのホスト名を取得する。
  domain_dns_record = regexall("\\S+", one(aws_amplify_domain_association.this.sub_domain).dns_record)
}

resource "aws_route53_record" "frontend_certificate_validation" {
  zone_id = var.hosted_zone_id
  name    = local.certificate_dns_record[0]
  type    = local.certificate_dns_record[1]
  records = [local.certificate_dns_record[2]]
  ttl     = 300

  # Amplifyが同じ検証レコードを自動作成した場合も、Terraform管理へ引き継ぐ。
  allow_overwrite = true
}

resource "aws_route53_record" "frontend" {
  zone_id = var.hosted_zone_id
  name    = var.domain_name
  type    = "A"

  # Amplify側が同じ配信用レコードを自動作成する場合にも対応する。
  allow_overwrite = true

  alias {
    name                   = local.domain_dns_record[1]
    zone_id                = "Z2FDTNDATAQYW2" # CloudFront共通のホストゾーンID
    evaluate_target_health = false
  }
}
