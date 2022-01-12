variable "system_domain" {
  type = string
}

variable "parent_zone" {
  type        = string
  default     = ""
  description = "The name of the parent zone for the provided system domain if it exists."
}

data "aws_route53_zone" "parent" {
  count = var.parent_zone == "" ? 0 : 1

  name = var.parent_zone
}

output "env_dns_zone_name_servers" {
  value = ["${split(",", local.name_servers)}"]
}

locals {
  zone_id      = var.parent_zone == "" ? element(concat(aws_route53_zone.env_dns_zone.*.zone_id, list("")), 0) : element(concat(data.aws_route53_zone.parent.*.zone_id, list("")), 0)
  name_servers = var.parent_zone == "" ? join(",", flatten(concat(aws_route53_zone.env_dns_zone.*.name_servers, list(list(""))))) : join(",", flatten(concat(data.aws_route53_zone.env_dns_zone.*.name_servers, list(list("")))))
}

resource "aws_route53_zone" "env_dns_zone" {
  count = var.parent_zone == "" ? 1 : 0

  name = var.system_domain

  tags = {
    Name = "${var.env_id}-hosted-zone"
  }
}

resource "aws_route53_record" "dns" {
  count = var.parent_zone == "" ? 1 : 0

  zone_id = local.zone_id
  name    = var.system_domain
  type    = "NS"
  ttl     = 300

  records = ["${local.name_servers}"]
}

resource "aws_route53_record" "wildcard_dns" {
  zone_id = local.zone_id
  name    = "*.${var.system_domain}"
  type    = "CNAME"
  ttl     = 300

  records = ["${aws_elb.cf_router_lb.dns_name}"]
}

resource "aws_route53_record" "ssh" {
  zone_id = local.zone_id
  name    = "ssh.${var.system_domain}"
  type    = "CNAME"
  ttl     = 300

  records = ["${aws_elb.cf_ssh_lb.dns_name}"]
}

resource "aws_route53_record" "bosh" {
  zone_id = local.zone_id
  name    = "bosh.${var.system_domain}"
  type    = "A"
  ttl     = 300

  records = ["${aws_eip.jumpbox_eip.public_ip}"]
}

resource "aws_route53_record" "tcp" {
  zone_id = local.zone_id
  name    = "tcp.${var.system_domain}"
  type    = "CNAME"
  ttl     = 300

  records = ["${aws_elb.cf_tcp_lb.dns_name}"]
}

resource "aws_route53_record" "iso" {
  count = var.isolation_segments

  zone_id = local.zone_id
  name    = "*.iso-seg.${var.system_domain}"
  type    = "CNAME"
  ttl     = 300

  records = ["${aws_elb.iso_router_lb.dns_name}"]
}
