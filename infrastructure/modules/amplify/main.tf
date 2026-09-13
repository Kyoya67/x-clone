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
