variable "kubernetes_master_host" {
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

resource "aws_route53_record" "api" {
  zone_id = "${aws_route53_zone.env_dns_zone.id}"
  name    = "${var.kubernetes_master_host}"
  type    = "CNAME"
  ttl     = 300

  records = ["${aws_elb.api.dns_name}"]
}
