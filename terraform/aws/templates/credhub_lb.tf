resource "aws_security_group" "credhub" {
  name        = "${var.env_id}-credhub-lb"
  description = "Credhub"
  vpc_id      = "${local.vpc_id}"

  tags {
    Name = "${var.env_id}-credhub-lb"
  }
}

resource "aws_security_group_rule" "credhub" {
  type        = "ingress"
  protocol    = "tcp"
  from_port   = 8844
  to_port     = 8844
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = "${aws_security_group.credhub.id}"
}

resource "aws_lb" "credhub" {
  name               = "${var.short_env_id}-credhub"
  load_balancer_type = "network"
  subnets            = ["${aws_subnet.lb_subnets.*.id}"]
}

resource "aws_lb_listener" "credhub" {
  load_balancer_arn = "${aws_lb.credhub.arn}"
  protocol          = "TCP"
  port              = 8844

  default_action {
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.credhub.arn}"
  }
}

resource "aws_lb_target_group" "credhub" {
  name     = "${var.short_env_id}-credhub"
  port     = 8844
  protocol = "TCP"
  vpc_id   = "${local.vpc_id}"

  health_check {
    healthy_threshold   = 10
    unhealthy_threshold = 10
    interval            = 30
    protocol            = "TCP"
  }
}

output "cf_credhub_lb_internal_security_group" {
  value = "${aws_security_group.credhub.name}"
}

output "cf_credhub_lb_target_groups" {
  value = ["${aws_lb_target_group.credhub.name}"]
}

output "cf_credhub_lb_name" {
  value = "${aws_lb.credhub.name}"
}

output "cf_credhub_lb_url" {
  value = "${aws_lb.credhub.dns_name}"
}
