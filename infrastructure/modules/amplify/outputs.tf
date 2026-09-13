output "app_id" {
  value = aws_amplify_app.this.id
}

output "branch_name" {
  value = aws_amplify_branch.this.branch_name
}

output "url" {
  value = "https://${aws_amplify_branch.this.branch_name}.${aws_amplify_app.this.default_domain}"
}

output "amplify_app_id" {
  value = module.amplify.app_id
}

output "amplify_url" {
  value = module.amplify.url
}
