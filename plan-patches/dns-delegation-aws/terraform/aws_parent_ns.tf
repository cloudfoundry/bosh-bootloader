variable "parent_zone_id" {
  type        = string
  description = "The AWS Route53 hosted zone ID for the 'parent' of the zone that bbl will create, used to set up DNS delegation"
}

resource "aws_route53_record" "perf-test" {
  name            = "${var.system_domain}"
  ttl             = 172800
  type            = "NS"
  zone_id         = "${var.parent_zone_id}"

  records = ["${aws_route53_zone.env_dns_zone.name_servers}"]
}