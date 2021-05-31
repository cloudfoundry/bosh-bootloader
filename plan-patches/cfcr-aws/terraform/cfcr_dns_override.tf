variable "kubernetes_master_host" {
  type = string
}

resource "aws_route53_zone" "cfcr_dns_zone" {
  name = "${var.kubernetes_master_host}"

  tags {
    Name = "${var.env_id}-cfcr-hosted-zone"
  }
}

resource "aws_route53_record" "cfcr_api" {
  zone_id = "${aws_route53_zone.cfcr_dns_zone.id}"
  name    = "api.${var.kubernetes_master_host}"
  type    = "CNAME"
  ttl     = 300

  records = ["${aws_elb.cfcr_api.dns_name}"]
}
