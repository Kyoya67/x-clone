output "issuer" {
  value = "https://cognito-idp.${data.aws_region.current.region}.amazonaws.com/${aws_cognito_user_pool.this.id}"
}

output "authorize_url" {
  value = "https://${aws_cognito_user_pool_domain.this.domain}.auth.${data.aws_region.current.region}.amazoncognito.com/oauth2/authorize"
}

output "token_url" {
  value = "https://${aws_cognito_user_pool_domain.this.domain}.auth.${data.aws_region.current.region}.amazoncognito.com/oauth2/token"
}

output "client_id" {
  value = aws_cognito_user_pool_client.this.id
}

output "client_secret" {
  value     = aws_cognito_user_pool_client.this.client_secret
  sensitive = true
}

data "aws_region" "current" {}

