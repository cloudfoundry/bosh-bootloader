resource "aws_lb" "alb_router" {
  name               = "alb-router"
  load_balancer_type = "application"

  security_groups = ["${aws_security_group.cf_router_lb_security_group.id}"]
  subnets         = ["${aws_subnet.lb_subnets.*.id}"]
}

resource "aws_lb_listener" "alb_router_443" {
  load_balancer_arn = "${aws_lb.alb_router.arn}"
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2015-05"
  certificate_arn   = "${aws_iam_server_certificate.lb_cert.arn}"

  default_action {
    target_group_arn = "${aws_lb_target_group.alb_router_443.arn}"
    type             = "forward"
  }
}

resource "aws_lb_target_group" "alb_router_443" {
  name     = "alb-router-target-443"
  port     = 443
  protocol = "HTTPS"
  vpc_id   = "${local.vpc_id}"

  health_check {
    path = "/health"
    port = 8080
  }
}

resource "aws_lb_listener" "alb_router_4443" {
  load_balancer_arn = "${aws_lb.alb_router.arn}"
  port              = "4443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2015-05"
  certificate_arn   = "${aws_iam_server_certificate.lb_cert.arn}"

  default_action {
    target_group_arn = "${aws_lb_target_group.alb_router_4443.arn}"
    type             = "forward"
  }
}

resource "aws_lb_target_group" "alb_router_4443" {
  name     = "alb-router-target-4443"
  port     = 4443
  protocol = "HTTPS"
  vpc_id   = "${local.vpc_id}"

  health_check {
    path = "/health"
    port = 8080
  }
}

resource "aws_lb_listener" "alb_router_80" {
  load_balancer_arn = "${aws_lb.alb_router.arn}"
  port              = "80"
  protocol          = "HTTP"

  default_action {
    target_group_arn = "${aws_lb_target_group.alb_router_80.arn}"
    type             = "forward"
  }
}

resource "aws_lb_target_group" "alb_router_80" {
  name     = "alb-router-target-80"
  port     = 80
  protocol = "HTTP"
  vpc_id   = "${local.vpc_id}"

  health_check {
    path = "/health"
    port = 8080
  }
}

resource "aws_elb" "cf_router_lb" {
  count = 0
  name  = "${var.short_env_id}-cf-router-lb"
}

output "cf_router_lb_name" {
  value = "${aws_lb.alb_router.name}"
}

output "cf_router_lb_url" {
  value = "${aws_lb.alb_router.dns_name}"
}

resource "aws_route53_record" "wildcard_dns" {
  zone_id = "${aws_route53_zone.env_dns_zone.id}"
  name    = "*.${var.system_domain}"
  type    = "CNAME"
  ttl     = 300

  records = ["${aws_lb.alb_router.dns_name}"]
}

output "cf_router_lb_target_groups" {
  value = ["${aws_lb_target_group.alb_router_443.name}", "${aws_lb_target_group.alb_router_4443.name}", "${aws_lb_target_group.alb_router_80.name}"]
}
