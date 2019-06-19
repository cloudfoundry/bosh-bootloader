# CF SSH
resource "aws_security_group" "cf_ssh" {
  name        = "${var.env_id}-cf-ssh-lb-security-group"
  description = "CF SSH Internal"
  vpc_id      = "${local.vpc_id}"

  ingress {
    cidr_blocks = ["0.0.0.0/0"]
    protocol    = "tcp"
    from_port   = 2222
    to_port     = 2222
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    Name = "${var.env_id}-cf-ssh-lb-security-group"
  }

  lifecycle {
    ignore_changes = ["name"]
  }
}

output "cf_ssh_security_group" {
  value = "${aws_security_group.cf_ssh.id}"
}

resource "aws_lb" "cf_ssh" {
  name                             = "${var.short_env_id}-cf-ssh-lb"
  load_balancer_type               = "network"
  enable_cross_zone_load_balancing = true
  internal                         = false
  subnets                          = ["${aws_subnet.lb_subnets.*.id}"]
}

resource "aws_lb_listener" "cf_ssh" {
  load_balancer_arn = "${aws_lb.cf_ssh.arn}"
  port              = 2222
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.cf_ssh.arn}"
  }
}

resource "aws_lb_target_group" "cf_ssh" {
  name     = "${var.short_env_id}-cf-ssh-lb"
  port     = 2222
  protocol = "TCP"
  vpc_id   = "${local.vpc_id}"

  health_check {
    protocol = "TCP"
  }
}

output "cf_ssh_lb_name" {
  value = "${aws_lb.cf_ssh.name}"
}

output "cf_ssh_lb_url" {
  value = "${aws_lb.cf_ssh.dns_name}"
}

output "cf_ssh_target_group_names" {
  value = ["${aws_lb_target_group.cf_ssh.name}"]
}

# CF Router

resource "aws_security_group" "cf_router" {
  name        = "${var.env_id}-cf-router-lb-security-group"
  description = "CF Router Internal"
  vpc_id      = "${local.vpc_id}"

  ingress {
    cidr_blocks = ["0.0.0.0/0"]
    protocol    = "tcp"
    from_port   = 80
    to_port     = 80
  }

  ingress {
    cidr_blocks = ["0.0.0.0/0"]
    protocol    = "tcp"
    from_port   = 443
    to_port     = 443
  }

  ingress {
    cidr_blocks = ["0.0.0.0/0"]
    protocol    = "tcp"
    from_port   = 4443
    to_port     = 4443
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    Name = "${var.env_id}-cf-router-lb-security-group"
  }

  lifecycle {
    ignore_changes = ["name"]
  }
}

output "cf_router_security_group" {
  value = "${aws_security_group.cf_router.id}"
}

resource "aws_lb" "cf_router" {
  name                             = "${var.short_env_id}-cf-router-lb"
  load_balancer_type               = "network"
  enable_cross_zone_load_balancing = true
  internal                         = false
  subnets                          = ["${aws_subnet.lb_subnets.*.id}"]
}

resource "aws_lb_listener" "cf_router_80" {
  load_balancer_arn = "${aws_lb.cf_router.arn}"
  port              = 80
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.cf_router_80.arn}"
  }
}

resource "aws_lb_listener" "cf_router_443" {
  load_balancer_arn = "${aws_lb.cf_router.arn}"
  port              = 443
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.cf_router_443.arn}"
  }
}

resource "aws_lb_listener" "cf_router_4443" {
  load_balancer_arn = "${aws_lb.cf_router.arn}"
  port              = 4443
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.cf_router_4443.arn}"
  }
}

resource "aws_lb_target_group" "cf_router_80" {
  name     = "${var.short_env_id}-routertg-80"
  port     = 80
  protocol = "TCP"
  vpc_id   = "${local.vpc_id}"

  health_check {
    protocol = "TCP"
  }
}

resource "aws_lb_target_group" "cf_router_443" {
  name     = "${var.short_env_id}-routertg-443"
  port     = 443
  protocol = "TCP"
  vpc_id   = "${local.vpc_id}"

  health_check {
    protocol = "TCP"
  }
}

resource "aws_lb_target_group" "cf_router_4443" {
  name     = "${var.short_env_id}-routertg-4443"
  port     = 4443
  protocol = "TCP"
  vpc_id   = "${local.vpc_id}"

  health_check {
    protocol = "TCP"
  }
}

output "cf_router_lb_name" {
  value = "${aws_lb.cf_router.name}"
}

output "cf_router_lb_url" {
  value = "${aws_lb.cf_router.dns_name}"
}

output "cf_router_target_group_names" {
  value = [
    "${aws_lb_target_group.cf_router_80.name}",
    "${aws_lb_target_group.cf_router_443.name}",
    "${aws_lb_target_group.cf_router_4443.name}",
  ]
}

# CF TCP Router

resource "aws_security_group" "cf_tcp_router" {
  name        = "${var.env_id}-cf-tcp-lb-security-group"
  description = "CF TCP Internal"
  vpc_id      = "${local.vpc_id}"

  ingress {
    cidr_blocks = ["0.0.0.0/0"]
    protocol    = "tcp"
    from_port   = 1024
    to_port     = 1033
  }

  ingress {
    cidr_blocks = ["0.0.0.0/0"]
    protocol    = "tcp"
    from_port   = 80
    to_port     = 80
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    Name = "${var.env_id}-cf-tcp-lb-security-group"
  }

  lifecycle {
    ignore_changes = ["name"]
  }
}

output "cf_tcp_router_security_group" {
  value = "${aws_security_group.cf_tcp_router.id}"
}

locals {
  tcp_router_listeners_count = 10
}

resource "aws_lb" "cf_tcp_router" {
  name                             = "${var.short_env_id}-cf-tcp-lb"
  load_balancer_type               = "network"
  enable_cross_zone_load_balancing = true
  internal                         = false
  subnets                          = ["${aws_subnet.lb_subnets.*.id}"]
}

resource "aws_lb_listener" "cf_tcp_router" {
  load_balancer_arn = "${aws_lb.cf_tcp_router.arn}"
  port              = "${1024 + count.index}"
  protocol          = "TCP"

  count = "${local.tcp_router_listeners_count}"

  default_action {
    type             = "forward"
    target_group_arn = "${element(aws_lb_target_group.cf_tcp_router.*.arn, count.index)}"
  }
}

resource "aws_lb_target_group" "cf_tcp_router" {
  name     = "${var.short_env_id}-tcptg-${1024 + count.index}"
  port     = "${1024 + count.index}"
  protocol = "TCP"
  vpc_id   = "${local.vpc_id}"

  count = "${local.tcp_router_listeners_count}"

  health_check {
    protocol = "TCP"
  }
}

output "cf_tcp_lb_name" {
  value = "${aws_lb.cf_tcp_router.name}"
}

output "cf_tcp_lb_url" {
  value = "${aws_lb.cf_tcp_router.dns_name}"
}

output "cf_tcp_target_group_names" {
  value = "${aws_lb_target_group.cf_tcp_router.*.name}"
}
