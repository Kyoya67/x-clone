resource "aws_cognito_user_pool" "this" {
  name = var.name

  username_attributes      = ["email"]
  auto_verified_attributes = ["email"]

  tags = var.tags
}

resource "aws_cognito_identity_provider" "google" {
  user_pool_id  = aws_cognito_user_pool.this.id
  provider_name = "Google"
  provider_type = "Google"

  provider_details = {
    client_id        = var.google_client_id
    client_secret    = var.google_client_secret
    authorize_scopes = "openid email profile"
  }

  attribute_mapping = {
    email    = "email"
    username = "sub"
    name     = "name"
  }
}

resource "aws_cognito_user_pool_client" "this" {
  name         = var.name
  user_pool_id = aws_cognito_user_pool.this.id

  supported_identity_providers         = [aws_cognito_identity_provider.google.provider_name]
  callback_urls                        = var.callback_urls
  logout_urls                          = var.logout_urls
  allowed_oauth_flows                  = ["code"]
  allowed_oauth_scopes                 = ["openid", "email", "profile"]
  allowed_oauth_flows_user_pool_client = true
  generate_secret                      = true
  prevent_user_existence_errors        = "ENABLED"

  explicit_auth_flows = [
    "ALLOW_REFRESH_TOKEN_AUTH",
  ]
}

resource "aws_cognito_user_pool_domain" "this" {
  domain       = var.domain_prefix
  user_pool_id = aws_cognito_user_pool.this.id
}

