resource "aws_lb" "cf_router_alb" {
  name               = "${var.short_env_id}-cf-router-lb"
  load_balancer_type = "application"

  security_groups = ["${aws_security_group.cf_router_alb_security_group.id}"]
  subnets         = flatten(["${aws_subnet.lb_subnets.*.id}"])
}

resource "aws_lb_listener" "cf_router_alb_443" {
  load_balancer_arn = "${aws_lb.cf_router_alb.arn}"
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS13-1-2-2021-06"
  certificate_arn   = "${aws_iam_server_certificate.lb_cert.arn}"

  default_action {
    target_group_arn = "${aws_lb_target_group.cf_router_alb_80.arn}"
    type             = "forward"
  }
}

resource "aws_lb_listener" "cf_router_alb_4443" {
  load_balancer_arn = "${aws_lb.cf_router_alb.arn}"
  port              = "4443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS13-1-2-2021-06"
  certificate_arn   = "${aws_iam_server_certificate.lb_cert.arn}"

  default_action {
    target_group_arn = "${aws_lb_target_group.cf_router_alb_80.arn}"
    type             = "forward"
  }
}

resource "aws_lb_listener" "cf_router_alb_80" {
  load_balancer_arn = "${aws_lb.cf_router_alb.arn}"
  port              = "80"
  protocol          = "HTTP"

  default_action {
    target_group_arn = "${aws_lb_target_group.cf_router_alb_80.arn}"
    type             = "forward"
  }
}

resource "aws_lb_target_group" "cf_router_alb_80" {
  name     = "${var.short_env_id}-routertg-80"
  port     = 80
  protocol = "HTTP"
  vpc_id   = "${local.vpc_id}"

  health_check {
    path = "/health"
    port = 8080
  }
}

resource "aws_security_group" "cf_router_alb_security_group" {
  name        = "${var.env_id}-cf-router-alb-security-group"
  description = "CF Router"
  vpc_id      = "${local.vpc_id}"

  ingress {
    cidr_blocks = ["0.0.0.0/0"]
    protocol        = "tcp"
    from_port       = 80
    to_port         = 80
  }

  ingress {
    cidr_blocks = ["0.0.0.0/0"]
    protocol        = "tcp"
    from_port       = 443
    to_port         = 443
  }

  ingress {
    cidr_blocks = ["0.0.0.0/0"]
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

  tags = {
    Name = "${var.env_id}-cf-router-alb-security-group"
  }

  lifecycle {
    ignore_changes = ["name"]
  }
}

resource "aws_security_group" "cf_router_alb_internal_security_group" {
  name        = "${var.env_id}-cf-router-alb-internal-security-group"
  description = "CF Router Internal"
  vpc_id      = "${local.vpc_id}"

  ingress {
    security_groups = ["${aws_security_group.cf_router_alb_security_group.id}"]
    protocol        = "tcp"
    from_port       = 80
    to_port         = 80
  }

  # Enable gorouter healthcheck
  ingress {
    security_groups = ["${aws_security_group.cf_router_alb_security_group.id}"]
    protocol        = "tcp"
    from_port       = 8080
    to_port         = 8080
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.env_id}-cf-router-alb-internal-security-group"
  }

  lifecycle {
    ignore_changes = [name]
  }
}

output "cf_router_alb_internal_security_group" {
  value = "${aws_security_group.cf_router_alb_internal_security_group.id}"
}

output "cf_router_alb_target_group" {
  value = "${aws_lb_target_group.cf_router_alb_80.name}"
}

output "cf_router_alb_name" {
  value = "${aws_lb.cf_router_alb.name}"
}

output "cf_router_alb_url" {
  value = "${aws_lb.cf_router_alb.dns_name}"
}
