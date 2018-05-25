resource "aws_subnet" "prom_lb_subnets" {
  count             = "${length(var.availability_zones)}"
  vpc_id            = "${local.vpc_id}"
  cidr_block        = "${cidrsubnet(var.vpc_cidr, 8, count.index+5)}"
  availability_zone = "${element(var.availability_zones, count.index)}"

  tags {
    Name = "${var.env_id}-lb-subnet${count.index}"
  }

  lifecycle {
    ignore_changes = ["cidr_block", "availability_zone"]
  }
}

resource "aws_route_table" "prom_lb_route_table" {
  vpc_id = "${local.vpc_id}"
}

resource "aws_route" "prom_lb_route_table" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.ig.id}"
  route_table_id         = "${aws_route_table.prom_lb_route_table.id}"
}

resource "aws_route_table_association" "route_prom_lb_subnets" {
  count          = "${length(var.availability_zones)}"
  subnet_id      = "${element(aws_subnet.prom_lb_subnets.*.id, count.index)}"
  route_table_id = "${aws_route_table.prom_lb_route_table.id}"
}

output "prom_lb_subnet_ids" {
  value = ["${aws_subnet.prom_lb_subnets.*.id}"]
}

output "prom_lb_subnet_availability_zones" {
  value = ["${aws_subnet.prom_lb_subnets.*.availability_zone}"]
}

output "prom_lb_subnet_cidrs" {
  value = ["${aws_subnet.prom_lb_subnets.*.cidr_block}"]
}

resource "aws_security_group" "prometheus_lb_internal_security_group" {
  name        = "${var.env_id}-prometheus-lb-internal-security-group"
  description = "Prometheus Internal"
  vpc_id      = "${local.vpc_id}"

  tags {
    Name = "${var.env_id}-prometheus-lb-internal-security-group"
  }

  lifecycle {
    ignore_changes = ["name"]
  }
}

resource "aws_security_group_rule" "prometheus_lb_internal_3000" {
  type        = "ingress"
  protocol    = "tcp"
  from_port   = 3000
  to_port     = 3000
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = "${aws_security_group.prometheus_lb_internal_security_group.id}"
}

resource "aws_security_group_rule" "prometheus_lb_internal_9090" {
  type        = "ingress"
  protocol    = "tcp"
  from_port   = 9090
  to_port     = 9090
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = "${aws_security_group.prometheus_lb_internal_security_group.id}"
}

resource "aws_security_group_rule" "prometheus_lb_internal_9093" {
  type        = "ingress"
  protocol    = "tcp"
  from_port   = 9093
  to_port     = 9093
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = "${aws_security_group.prometheus_lb_internal_security_group.id}"
}

resource "aws_security_group_rule" "prometheus_lb_internal_egress" {
  type        = "egress"
  protocol    = "-1"
  from_port   = 0
  to_port     = 0
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = "${aws_security_group.prometheus_lb_internal_security_group.id}"
}

resource "aws_lb" "prometheus_lb" {
  name               = "${var.short_env_id}-prometheus-lb"
  load_balancer_type = "network"
  subnets            = ["${aws_subnet.prom_lb_subnets.*.id}"]
}

resource "aws_lb_listener" "prometheus_lb_3000" {
  load_balancer_arn = "${aws_lb.prometheus_lb.arn}"
  protocol          = "TCP"
  port              = 3000

  default_action {
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.prometheus_lb_3000.arn}"
  }
}

resource "aws_lb_target_group" "prometheus_lb_3000" {
  name     = "${var.short_env_id}-prometheus3000"
  port     = 3000
  protocol = "TCP"
  vpc_id   = "${local.vpc_id}"

  health_check {
    healthy_threshold   = 10
    unhealthy_threshold = 10
    interval            = 30
    protocol            = "TCP"
  }
}

resource "aws_lb_listener" "prometheus_lb_9090" {
  load_balancer_arn = "${aws_lb.prometheus_lb.arn}"
  protocol          = "TCP"
  port              = 9090

  default_action {
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.prometheus_lb_9090.arn}"
  }
}

resource "aws_lb_listener" "prometheus_lb_9093" {
  load_balancer_arn = "${aws_lb.prometheus_lb.arn}"
  protocol          = "TCP"
  port              = 9093

  default_action {
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.prometheus_lb_9093.arn}"
  }
}

resource "aws_lb_target_group" "prometheus_lb_9090" {
  name     = "${var.short_env_id}-prometheus9090"
  port     = 9090
  protocol = "TCP"
  vpc_id   = "${local.vpc_id}"
}

resource "aws_lb_target_group" "prometheus_lb_9093" {
  name     = "${var.short_env_id}-prometheus9093"
  port     = 9093
  protocol = "TCP"
  vpc_id   = "${local.vpc_id}"
}

output "prometheus_lb_internal_security_group" {
  value = "${aws_security_group.prometheus_lb_internal_security_group.name}"
}

output "prometheus_lb_target_groups" {
  value = ["${aws_lb_target_group.prometheus_lb_3000.name}", "${aws_lb_target_group.prometheus_lb_9090.name}", "${aws_lb_target_group.prometheus_lb_9093.name}"]
}

output "prometheus_lb_name" {
  value = "${aws_lb.prometheus_lb.name}"
}

output "prometheus_lb_url" {
  value = "${aws_lb.prometheus_lb.dns_name}"
}
