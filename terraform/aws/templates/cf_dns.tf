variable "system_domain" {
  type = "string"
}

resource "aws_route53_zone" "env_dns_zone" {
  name = "${var.system_domain}"

  tags {
    Name = "${var.env_id}-hosted-zone"
  }
}

output "env_dns_zone_name_servers" {
  value = "${aws_route53_zone.env_dns_zone.name_servers}"
}

resource "aws_route53_record" "wildcard_dns" {
  zone_id = "${aws_route53_zone.env_dns_zone.id}"
  name    = "*.${var.system_domain}"
  type    = "CNAME"
  ttl     = 300

  records = ["${aws_elb.cf_router_lb.dns_name}"]
}

resource "aws_route53_record" "ssh" {
  zone_id = "${aws_route53_zone.env_dns_zone.id}"
  name    = "ssh.${var.system_domain}"
  type    = "CNAME"
  ttl     = 300

  records = ["${aws_elb.cf_ssh_lb.dns_name}"]
}

resource "aws_route53_record" "bosh" {
  zone_id = "${aws_route53_zone.env_dns_zone.id}"
  name    = "bosh.${var.system_domain}"
  type    = "A"
  ttl     = 300

  records = ["${aws_eip.jumpbox_eip.public_ip}"]
}

resource "aws_route53_record" "tcp" {
  zone_id = "${aws_route53_zone.env_dns_zone.id}"
  name    = "tcp.${var.system_domain}"
  type    = "CNAME"
  ttl     = 300

  records = ["${aws_elb.cf_tcp_lb.dns_name}"]
}

resource "aws_route53_record" "iso" {
  count = "${var.isolation_segments}"

  zone_id = "${aws_route53_zone.env_dns_zone.id}"
  name    = "*.iso-seg.${var.system_domain}"
  type    = "CNAME"
  ttl     = 300

  records = ["${aws_elb.iso_router_lb.dns_name}"]
}

resource "aws_route53_record" "credhub" {
  zone_id = "${aws_route53_zone.env_dns_zone.id}"
  name    = "credhub.${var.system_domain}"
  type    = "CNAME"
  ttl     = 300

  records = ["${aws_lb.credhub.dns_name}"]
}
