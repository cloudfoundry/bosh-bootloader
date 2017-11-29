resource "aws_security_group" "concourse_lb_internal_security_group" {
  description = "Concourse Internal"
  vpc_id      = "${aws_vpc.vpc.id}"

  tags {
    Name = "${var.env_id}-concourse-lb-internal-security-group"
  }
}

resource "aws_security_group_rule" "concourse_lb_internal_80" {
  type        = "ingress"
  protocol    = "tcp"
  from_port   = 80
  to_port     = 80
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = "${aws_security_group.concourse_lb_internal_security_group.id}"
}

resource "aws_security_group_rule" "concourse_lb_internal_2222" {
  type        = "ingress"
  protocol    = "tcp"
  from_port   = 2222
  to_port     = 2222
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = "${aws_security_group.concourse_lb_internal_security_group.id}"
}

resource "aws_security_group_rule" "concourse_lb_internal_443" {
  type        = "ingress"
  protocol    = "tcp"
  from_port   = 443
  to_port     = 443
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = "${aws_security_group.concourse_lb_internal_security_group.id}"
}

resource "aws_security_group_rule" "concourse_lb_internal_egress" {
  type        = "egress"
  protocol    = "-1"
  from_port   = 0
  to_port     = 0
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = "${aws_security_group.concourse_lb_internal_security_group.id}"
}

resource "aws_lb" "concourse_lb" {
  name               = "${var.short_env_id}-concourse-lb"
  load_balancer_type = "network"
  subnets            = ["${aws_subnet.lb_subnets.*.id}"]
}

resource "aws_lb_listener" "concourse_lb_80" {
  load_balancer_arn = "${aws_lb.concourse_lb.arn}"
  protocol          = "TCP"
  port              = 80

  default_action {
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.concourse_lb_80.arn}"
  }
}

resource "aws_lb_target_group" "concourse_lb_80" {
  name     = "${var.short_env_id}-concourse80"
  port     = 80
  protocol = "TCP"
  vpc_id   = "${aws_vpc.vpc.id}"

  health_check {
    healthy_threshold   = 10
    unhealthy_threshold = 10
    interval            = 30
    timeout             = 5
    protocol            = "TCP"
  }
}

resource "aws_lb_listener" "concourse_lb_2222" {
  load_balancer_arn = "${aws_lb.concourse_lb.arn}"
  protocol          = "TCP"
  port              = 2222

  default_action {
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.concourse_lb_2222.arn}"
  }
}

resource "aws_lb_target_group" "concourse_lb_2222" {
  name     = "${var.short_env_id}-concourse2222"
  port     = 2222
  protocol = "TCP"
  vpc_id   = "${aws_vpc.vpc.id}"
}

resource "aws_lb_listener" "concourse_lb_443" {
  load_balancer_arn = "${aws_lb.concourse_lb.arn}"
  protocol          = "TCP"
  port              = 443

  default_action {
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.concourse_lb_443.arn}"
  }
}

resource "aws_lb_target_group" "concourse_lb_443" {
  name     = "${var.short_env_id}-concourse443"
  port     = 443
  protocol = "TCP"
  vpc_id   = "${aws_vpc.vpc.id}"
}

output "concourse_lb_internal_security_group" {
  value = "${aws_security_group.concourse_lb_internal_security_group.name}"
}

output "concourse_lb_target_groups" {
  value = ["${var.short_env_id}-concourse-lb-80", "${var.short_env_id}-concourse-lb-443", "${var.short_env_id}-concourse-lb-2222"]
}

output "concourse_lb_name" {
  value = "${aws_lb.concourse_lb.name}"
}

output "concourse_lb_url" {
  value = "${aws_lb.concourse_lb.dns_name}"
}
