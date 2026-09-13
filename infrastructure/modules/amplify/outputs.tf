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
  value = aws_amplify_app.this.id
}

output "amplify_url" {
  value = "https://${aws_amplify_branch.this.branch_name}.${aws_amplify_app.this.default_domain}"
}

output "certificate_dns_record" {
  value = {
    name  = local.certificate_dns_record[0]
    type  = local.certificate_dns_record[1]
    value = local.certificate_dns_record[2]
  }
}

output "domain_dns_record" {
  value = {
    name  = var.domain_name
    value = local.domain_dns_record[1]
  }
}
