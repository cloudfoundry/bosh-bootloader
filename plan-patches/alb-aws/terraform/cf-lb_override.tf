resource "aws_lb" "cf_router" {
  name               = "${var.short_env_id}-cf-router-lb"
  load_balancer_type = "application"

  security_groups = ["${aws_security_group.cf_router_lb_security_group.id}"]
  subnets         = ["${aws_subnet.lb_subnets.*.id}"]
}

resource "aws_lb_listener" "cf_router_443" {
  load_balancer_arn = "${aws_lb.cf_router.arn}"
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2015-05"
  certificate_arn   = "${aws_iam_server_certificate.lb_cert.arn}"

  default_action {
    target_group_arn = "${aws_lb_target_group.cf_router_443.arn}"
    type             = "forward"
  }
}

resource "aws_lb_target_group" "cf_router_443" {
  name     = "${var.short_env_id}-routertg-443"
  port     = 443
  protocol = "HTTPS"
  vpc_id   = "${local.vpc_id}"

  health_check {
    path = "/health"
    port = 8080
  }
}

resource "aws_lb_listener" "cf_router_4443" {
  load_balancer_arn = "${aws_lb.cf_router.arn}"
  port              = "4443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2015-05"
  certificate_arn   = "${aws_iam_server_certificate.lb_cert.arn}"

  default_action {
    target_group_arn = "${aws_lb_target_group.cf_router_4443.arn}"
    type             = "forward"
  }
}

resource "aws_lb_target_group" "cf_router_4443" {
  name     = "${var.short_env_id}-routertg-4443"
  port     = 4443
  protocol = "HTTPS"
  vpc_id   = "${local.vpc_id}"

  health_check {
    path = "/health"
    port = 8080
  }
}

resource "aws_lb_listener" "cf_router_80" {
  load_balancer_arn = "${aws_lb.cf_router.arn}"
  port              = "80"
  protocol          = "HTTP"

  default_action {
    target_group_arn = "${aws_lb_target_group.cf_router_80.arn}"
    type             = "forward"
  }
}

resource "aws_lb_target_group" "cf_router_80" {
  name     = "${var.short_env_id}-routertg-80"
  port     = 80
  protocol = "HTTP"
  vpc_id   = "${local.vpc_id}"

  health_check {
    path = "/health"
    port = 8080
  }
}

resource "aws_security_group" "cf_router_lb_internal_security_group" {
  name        = "${var.env_id}-cf-router-lb-internal-security-group"
  description = "CF Router Internal"
  vpc_id      = "${local.vpc_id}"

  ingress {
    security_groups = ["${aws_security_group.cf_router_lb_security_group.id}"]
    protocol        = "tcp"
    from_port       = 80
    to_port         = 80
  }

  ingress {
    security_groups = ["${aws_security_group.cf_router_lb_security_group.id}"]
    protocol        = "tcp"
    from_port       = 8080
    to_port         = 8080
  }

  ingress {
    security_groups = ["${aws_security_group.cf_router_lb_security_group.id}"]
    protocol        = "tcp"
    from_port       = 443
    to_port         = 443
  }

  ingress {
    security_groups = ["${aws_security_group.cf_router_lb_security_group.id}"]
    protocol        = "tcp"
    from_port       = 4443
    to_port         = 4443
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    Name = "${var.env_id}-cf-router-lb-internal-security-group"
  }

  lifecycle {
    ignore_changes = ["name"]
  }
}

output "cf_router_lb_name" {
  value = "${aws_lb.cf_router.name}"
}

output "cf_router_lb_url" {
  value = "${aws_lb.cf_router.dns_name}"
}

resource "aws_route53_record" "wildcard_dns" {
  zone_id = "${aws_route53_zone.env_dns_zone.id}"
  name    = "*.${var.system_domain}"
  type    = "CNAME"
  ttl     = 300

  records = ["${aws_lb.cf_router.dns_name}"]
}
