resource "aws_subnet" "credhub_lb_subnets" {
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

resource "aws_route_table" "credhub_lb_route_table" {
  vpc_id = "${local.vpc_id}"
}

resource "aws_route" "credhub_lb_route_table" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.ig.id}"
  route_table_id         = "${aws_route_table.credhub_lb_route_table.id}"
}

resource "aws_route_table_association" "route_credhub_lb_subnets" {
  count          = "${length(var.availability_zones)}"
  subnet_id      = "${element(aws_subnet.credhub_lb_subnets.*.id, count.index)}"
  route_table_id = "${aws_route_table.credhub_lb_route_table.id}"
}

output "credhub_lb_subnet_ids" {
  value = ["${aws_subnet.credhub_lb_subnets.*.id}"]
}

output "credhub_lb_subnet_availability_zones" {
  value = ["${aws_subnet.credhub_lb_subnets.*.availability_zone}"]
}

output "credhub_lb_subnet_cidrs" {
  value = ["${aws_subnet.credhub_lb_subnets.*.cidr_block}"]
}

resource "aws_security_group" "credhub_lb_internal_security_group" {
  name        = "${var.env_id}-credhub-lb-internal-security-group"
  description = "Credhub Internal"
  vpc_id      = "${local.vpc_id}"

  tags {
    Name = "${var.env_id}-credhub-lb-internal-security-group"
  }

  lifecycle {
    ignore_changes = ["name"]
  }
}

resource "aws_security_group_rule" "credhub_lb_internal_8844" {
  type        = "ingress"
  protocol    = "tcp"
  from_port   = 8844
  to_port     = 8844
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = "${aws_security_group.credhub_lb_internal_security_group.id}"
}

resource "aws_security_group_rule" "credhub_lb_internal_egress" {
  type        = "egress"
  protocol    = "-1"
  from_port   = 0
  to_port     = 0
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = "${aws_security_group.credhub_lb_internal_security_group.id}"
}

output "credhub_lb_internal_security_group" {
  value = "${aws_security_group.credhub_lb_internal_security_group.name}"
}

resource "aws_lb" "credhub_lb" {
  name               = "${var.short_env_id}-credhub-lb"
  load_balancer_type = "network"
  subnets            = ["${aws_subnet.credhub_lb_subnets.*.id}"]
}

resource "aws_lb_listener" "credhub_lb_8844" {
  load_balancer_arn = "${aws_lb.credhub_lb.arn}"
  protocol          = "TCP"
  port              = 8844

  default_action {
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.credhub_lb_8844.arn}"
  }
}

resource "aws_lb_target_group" "credhub_lb_8844" {
  name     = "${var.short_env_id}-credhub8844"
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

output "credhub_lb_target_group" {
  value = ["${aws_lb_target_group.credhub_lb_8844.name}"]
}

output "credhub_lb_name" {
  value = "${aws_lb.credhub_lb.name}"
}

output "credhub_lb_url" {
  value = "${aws_lb.credhub_lb.dns_name}"
}
