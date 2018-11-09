variable "isolation_segments" {
  type        = "string"
  default     = "0"
  description = "Optionally create a load balancer and DNS entries for a single isolation segment. Valid values are 0 or 1."
}

variable "iso_to_bosh_ports" {
  type    = "list"
  default = [22, 6868, 2555, 4222, 25250]
}

variable "iso_to_shared_tcp_ports" {
  type    = "list"
  default = [9090, 9091, 8082, 8300, 8301, 8889, 8443, 3000, 4443, 8080, 3457, 9023, 9022, 4222]
}

variable "iso_to_shared_udp_ports" {
  type    = "list"
  default = [8301, 8302, 8600]
}

locals {
  iso_az_count = "${var.isolation_segments > 0 ? length(var.availability_zones) : 0}"
}

resource "aws_subnet" "iso_subnets" {
  count             = "${local.iso_az_count}"
  vpc_id            = "${local.vpc_id}"
  cidr_block        = "${cidrsubnet(var.vpc_cidr, 4, count.index + length(var.availability_zones) + 1)}"
  availability_zone = "${element(var.availability_zones, count.index)}"

  tags {
    Name = "${var.env_id}-iso-subnet${count.index}"
  }
}

resource "aws_route_table_association" "route_iso_subnets" {
  count          = "${local.iso_az_count}"
  subnet_id      = "${element(aws_subnet.iso_subnets.*.id, count.index)}"
  route_table_id = "${aws_route_table.internal_route_table.id}"
}

resource "aws_lb" "iso_router_lb" {
  count                            = "${var.isolation_segments}"
  name                             = "${var.short_env_id}-iso-router-lb"
  load_balancer_type               = "network"
  enable_cross_zone_load_balancing = true
  internal                         = false
  subnets                          = ["${aws_subnet.lb_subnets.*.id}"]
}

resource "aws_lb_listener" "iso_router_lb_80" {
  count = "${var.isolation_segments}"
  load_balancer_arn = "${aws_lb.iso_router_lb.arn}"
  port              = 80
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.iso_router_lb_80.arn}"
  }
}

resource "aws_lb_listener" "iso_router_lb_443" {
  count = "${var.isolation_segments}"
  load_balancer_arn = "${aws_lb.iso_router_lb.arn}"
  port              = 443
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.iso_router_lb_443.arn}"
  }
}

resource "aws_lb_listener" "iso_router_lb_4443" {
  count = "${var.isolation_segments}"
  load_balancer_arn = "${aws_lb.iso_router_lb.arn}"
  port              = 4443
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.iso_router_lb_4443.arn}"
  }
}

resource "aws_lb_target_group" "iso_router_lb_80" {
  count = "${var.isolation_segments}"
  name     = "${var.short_env_id}-isotg-80"
  port     = 80
  protocol = "TCP"
  vpc_id   = "${local.vpc_id}"

  health_check {
    protocol = "TCP"
  }
}

resource "aws_lb_target_group" "iso_router_lb_443" {
  count = "${var.isolation_segments}"
  name     = "${var.short_env_id}-isotg-443"
  port     = 443
  protocol = "TCP"
  vpc_id   = "${local.vpc_id}"

  health_check {
    protocol = "TCP"
  }
}

resource "aws_lb_target_group" "iso_router_lb_4443" {
  count = "${var.isolation_segments}"
  name     = "${var.short_env_id}-isotg-4443"
  port     = 4443
  protocol = "TCP"
  vpc_id   = "${local.vpc_id}"

  health_check {
    protocol = "TCP"
  }
}

resource "aws_security_group" "iso_security_group" {
  count = "${var.isolation_segments}"

  name   = "${var.env_id}-iso-sg"
  vpc_id = "${local.vpc_id}"

  description = "Private isolation segment"

  tags {
    Name = "${var.env_id}-iso-security-group"
  }
}

resource "aws_security_group" "iso_shared_security_group" {
  count = "${var.isolation_segments}"

  name   = "${var.env_id}-iso-shared-sg"
  vpc_id = "${local.vpc_id}"

  description = "Shared isolation segments"

  tags {
    Name = "${var.env_id}-iso-shared-security-group"
  }
}

resource "aws_security_group_rule" "isolation_segments_to_bosh_rule" {
  count = "${var.isolation_segments * length(var.iso_to_bosh_ports)}"

  description = "TCP traffic from iso-sg to bosh"

  security_group_id        = "${aws_security_group.bosh_security_group.id}"
  type                     = "ingress"
  protocol                 = "tcp"
  to_port                  = "${element(var.iso_to_bosh_ports, count.index)}"
  from_port                = "${element(var.iso_to_bosh_ports, count.index)}"
  source_security_group_id = "${aws_security_group.iso_security_group.id}"
}

resource "aws_security_group_rule" "isolation_segments_to_shared_tcp_rule" {
  count = "${var.isolation_segments * length(var.iso_to_shared_tcp_ports)}"

  description = "TCP traffic from iso-sg to iso-shared-sg"

  security_group_id        = "${aws_security_group.iso_shared_security_group.id}"
  type                     = "ingress"
  protocol                 = "tcp"
  to_port                  = "${element(var.iso_to_shared_tcp_ports, count.index)}"
  from_port                = "${element(var.iso_to_shared_tcp_ports, count.index)}"
  source_security_group_id = "${aws_security_group.iso_security_group.id}"
}

resource "aws_security_group_rule" "isolation_segments_to_shared_udp_rule" {
  count = "${var.isolation_segments * length(var.iso_to_shared_udp_ports)}"

  description = "UDP traffic from iso-sg to iso-shared-sg"

  security_group_id        = "${aws_security_group.iso_shared_security_group.id}"
  type                     = "ingress"
  protocol                 = "udp"
  to_port                  = "${element(var.iso_to_shared_udp_ports, count.index)}"
  from_port                = "${element(var.iso_to_shared_udp_ports, count.index)}"
  source_security_group_id = "${aws_security_group.iso_security_group.id}"
}

resource "aws_security_group_rule" "isolation_segments_to_bosh_all_traffic_rule" {
  count = "${var.isolation_segments}"

  description = "ALL traffic from iso-sg to bosh"

  depends_on               = ["aws_security_group.bosh_security_group"]
  security_group_id        = "${aws_security_group.bosh_security_group.id}"
  type                     = "ingress"
  protocol                 = "-1"
  from_port                = 0
  to_port                  = 0
  source_security_group_id = "${aws_security_group.iso_security_group.id}"
}

resource "aws_security_group_rule" "shared_diego_bbs_to_isolated_cells_rule" {
  count = "${var.isolation_segments}"

  description = "TCP traffic from shared diego bbs to iso-sg"

  depends_on               = ["aws_security_group.iso_security_group"]
  security_group_id        = "${aws_security_group.iso_security_group.id}"
  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 1801
  to_port                  = 1801
  source_security_group_id = "${aws_security_group.iso_shared_security_group.id}"
}

resource "aws_security_group_rule" "nat_to_isolated_cells_rule" {
  count = "${var.isolation_segments}"

  description = "ALL traffic from nat-sg to iso-sg"

  security_group_id        = "${aws_security_group.nat_security_group.id}"
  type                     = "ingress"
  protocol                 = "-1"
  from_port                = 0
  to_port                  = 0
  source_security_group_id = "${aws_security_group.iso_security_group.id}"
}

output "cf_iso_router_lb_name" {
  value = "${element(concat(aws_lb.iso_router_lb.*.name, list("")), 0)}"
}

output "cf_iso_router_target_group_names" {
  value = [
    "${element(concat(aws_lb_target_group.iso_router_lb_80.*.name, list("")), 0)}",
    "${element(concat(aws_lb_target_group.iso_router_lb_443.*.name, list("")), 0)}",
    "${element(concat(aws_lb_target_group.iso_router_lb_4443.*.name, list("")), 0)}",
  ]
}

output "iso_security_group_id" {
  value = "${element(concat(aws_security_group.iso_security_group.*.id, list("")), 0)}"
}

output "iso_az_subnet_id_mapping" {
  value = "${
    zipmap("${aws_subnet.iso_subnets.*.availability_zone}", "${aws_subnet.iso_subnets.*.id}")
  }"
}

output "iso_az_subnet_cidr_mapping" {
  value = "${
    zipmap("${aws_subnet.iso_subnets.*.availability_zone}", "${aws_subnet.iso_subnets.*.cidr_block}")
  }"
}

output "iso_shared_security_group_id" {
  value = "${element(concat(aws_security_group.iso_shared_security_group.*.id, list("")), 0)}"
}
