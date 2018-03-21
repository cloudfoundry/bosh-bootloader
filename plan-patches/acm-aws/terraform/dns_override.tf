resource "aws_route53_zone" "env_dns_zone" {
  count = 0
}

data "aws_route53_zone" "env_dns_zone" {
  name = "${var.system_domain}"
}

output "env_dns_zone_name_servers" {
  value = ""
}

resource "aws_route53_record" "wildcard_dns" {
  zone_id = "${data.aws_route53_zone.env_dns_zone.id}"
}

resource "aws_route53_record" "ssh" {
  zone_id = "${data.aws_route53_zone.env_dns_zone.id}"
}

resource "aws_route53_record" "bosh" {
  zone_id = "${data.aws_route53_zone.env_dns_zone.id}"
}

resource "aws_route53_record" "tcp" {
  zone_id = "${data.aws_route53_zone.env_dns_zone.id}"
}
