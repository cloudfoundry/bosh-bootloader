variable "system_domain" {
  type = "string"
}

data "aws_route53_zone" "env_dns_zone" {
  name = "${var.system_domain}"
}

resource "aws_route53_zone" "env_dns_zone" {
  name  = "${var.system_domain}"
  count = "${data.aws_route53_zone.env_dns_zone.zone_id == "" ? 1 : 0}"

  tags {
    Name = "${var.env_id}-hosted-zone"
  }
}

locals {
  zone_id                  = "${data.aws_route53_zone.env_dns_zone.zone_id == "" ? element(concat(aws_route53_zone.env_dns_zone.*.zone_id, list("")), 0) : data.aws_route53_zone.env_dns_zone.zone_id}"
  data_dns_nameservers     = "${join(",", data.aws_route53_zone.env_dns_zone.name_servers)}"
  resource_dns_nameservers = "${join(",", concat(aws_route53_zone.env_dns_zone.*.name_servers, list("")))}"
}

output "env_dns_zone_name_servers" {
  value = ["${split(",", data.aws_route53_zone.env_dns_zone.zone_id == "" ?  local.resource_dns_nameservers : local.data_dns_nameservers)}"]
}

resource "aws_route53_record" "wildcard_dns" {
  zone_id = "${local.zone_id}"
  name    = "*.${var.system_domain}"
  type    = "CNAME"
  ttl     = 300

  records = ["${aws_elb.cf_router_lb.dns_name}"]
}

resource "aws_route53_record" "ssh" {
  zone_id = "${local.zone_id}"
  name    = "ssh.${var.system_domain}"
  type    = "CNAME"
  ttl     = 300

  records = ["${aws_elb.cf_ssh_lb.dns_name}"]
}

resource "aws_route53_record" "bosh" {
  zone_id = "${local.zone_id}"
  name    = "bosh.${var.system_domain}"
  type    = "A"
  ttl     = 300

  records = ["${aws_eip.jumpbox_eip.public_ip}"]
}

resource "aws_route53_record" "tcp" {
  zone_id = "${local.zone_id}"
  name    = "tcp.${var.system_domain}"
  type    = "CNAME"
  ttl     = 300

  records = ["${aws_elb.cf_tcp_lb.dns_name}"]
}

resource "aws_route53_record" "iso" {
  count = "${var.isolation_segments}"

  zone_id = "${local.zone_id}"
  name    = "*.iso-seg.${var.system_domain}"
  type    = "CNAME"
  ttl     = 300

  records = ["${aws_elb.iso_router_lb.dns_name}"]
}
