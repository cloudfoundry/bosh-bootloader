variable "system_domain" {
  type = "string"
}

variable "existing_zone_id" {
  type    = "string"
  default = ""
}

variable "existing_zone_ns" {
  type    = "string"
  default = ""
}

resource "aws_route53_zone" "env_dns_zone" {
  name  = "${var.system_domain}"
  count = "${var.existing_zone_id == "" ? 1 : 0}"

  tags {
    Name = "${var.env_id}-hosted-zone"
  }
}

locals {
  zone_id     = "${var.existing_zone_id == "" ? element(concat(aws_route53_zone.env_dns_zone.*.zone_id, list("")), 0) : var.existing_zone_id}"
  new_zone_ns = "${join(",", concat(aws_route53_zone.env_dns_zone.*.name_servers, list("")))}"
}

output "env_dns_zone_name_servers" {
  value = ["${split(",", var.existing_zone_ns == "" ?  local.new_zone_ns : var.existing_zone_ns)}"]
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
