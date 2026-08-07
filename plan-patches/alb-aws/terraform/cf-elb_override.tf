# remap CNAME record from ELB to ALB DNS name
resource "aws_route53_record" "wildcard_dns" {
  zone_id = "${aws_route53_zone.env_dns_zone[0].id}"
  name    = "*.${var.system_domain}"
  type    = "CNAME"
  ttl     = 300

  records = ["${aws_lb.cf_router_alb.dns_name}"]
}
