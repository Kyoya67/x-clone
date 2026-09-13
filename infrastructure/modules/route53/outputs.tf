output "zone_id" {
  value = aws_route53_zone.this.zone_id
}

output "name_servers" {
  value = aws_route53_zone.this.name_servers
}

output "api_certificate_validation_record_fqdns" {
  value = [for record in aws_route53_record.api_certificate_validation : record.fqdn]
}

output "api_record_name" {
  value = try(aws_route53_record.api[0].name, null)
}

output "amplify_frontend_record_name" {
  value = try(aws_route53_record.amplify_frontend[0].name, null)
}
