variable "isolation_segments" {
  type        = string
  default     = "0"
  description = "Optionally create a load balancer and DNS entries for a single isolation segment. Valid values are 0 or 1."
}

variable "iso_to_bosh_ports" {
  type    = list(number)
  default = [22, 6868, 2555, 4222, 25250]
}

variable "iso_to_shared_tcp_ports" {
  type    = list(number)
  default = [9090, 9091, 8082, 8300, 8301, 8889, 8443, 3000, 4443, 8080, 3457, 9023, 9022, 4222]
}

variable "iso_to_shared_udp_ports" {
  type    = list(number)
  default = [8301, 8302, 8600]
}

locals {
  iso_az_count = var.isolation_segments > 0 ? length(var.availability_zones) : 0
}

resource "aws_subnet" "iso_subnets" {
  count                           = local.iso_az_count
  vpc_id                          = local.vpc_id
  cidr_block                      = cidrsubnet(var.vpc_cidr, 4, count.index + length(var.availability_zones) + 1)
  ipv6_cidr_block                 = var.dualstack ? "${cidrsubnet(aws_vpc.vpc[0].ipv6_cidr_block, 8, count.index + 2 + length(var.availability_zones))}" : null
  availability_zone               = element(var.availability_zones, count.index)
  assign_ipv6_address_on_creation = var.dualstack
  enable_dns64                    = var.dualstack

  tags = {
    Name = "${var.env_id}-iso-subnet${count.index}"
  }
}

resource "aws_route_table_association" "route_iso_subnets" {
  count          = local.iso_az_count
  subnet_id      = aws_subnet.iso_subnets[count.index].id
  route_table_id = aws_route_table.nated_route_table.id
}


resource "aws_elb" "iso_router_lb" {
  count = var.isolation_segments == "1" && var.dualstack == false ? 1 : 0

  name                      = "${var.short_env_id}-iso-router-lb"
  cross_zone_load_balancing = true

  health_check {
    healthy_threshold   = 5
    unhealthy_threshold = 2
    interval            = 12
    target              = "TCP:80"
    timeout             = 2
  }

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  listener {
    instance_port      = 80
    instance_protocol  = "http"
    lb_port            = 443
    lb_protocol        = "https"
    ssl_certificate_id = aws_iam_server_certificate.lb_cert.arn
  }

  listener {
    instance_port      = 80
    instance_protocol  = "tcp"
    lb_port            = 4443
    lb_protocol        = "ssl"
    ssl_certificate_id = aws_iam_server_certificate.lb_cert.arn
  }

  security_groups = ["${aws_security_group.cf_router_lb_security_group.id}"]
  subnets         = ["${aws_subnet.lb_subnets.*.id}"]

  tags = {
    Name = "${var.env_id}"
  }
}

resource "aws_lb" "iso_router_nlb" {
  count              = var.isolation_segments == "1" && var.dualstack ? 1 : 0
  name               = "${var.short_env_id}-iso-router-lb"
  internal           = false
  load_balancer_type = "network"
  security_groups    = [aws_security_group.cf_router_lb_security_group.id]
  subnets            = [for subnet in aws_subnet.lb_subnets : subnet.id]

  enable_deletion_protection       = false
  enable_cross_zone_load_balancing = true

  # idle_timeout = var.elb_idle_timeout
  ip_address_type = "dualstack"

  tags = {
    Name = var.env_id
  }
}

resource "aws_lb_target_group" "iso_router_nlb_http" {
  count    = var.isolation_segments == "1" && var.dualstack ? 1 : 0
  name     = "${var.short_env_id}-iso-router-nlb-http"
  port     = 80
  protocol = "HTTP"
  vpc_id   = local.vpc_id

  health_check {
    healthy_threshold   = 5
    unhealthy_threshold = 2
    interval            = 15
    protocol            = "TCP"
    port                = 80
  }

  tags = {
    Name = "${var.env_id}"
  }
}

resource "aws_lb_listener" "iso_router_nlb_http" {
  count             = var.isolation_segments == "1" && var.dualstack ? 1 : 0
  load_balancer_arn = aws_lb.iso_router_nlb[0].arn
  port              = "80"
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.iso_router_nlb_http[0].arn
  }
}

resource "aws_lb_listener" "iso_router_nlb_https" {
  count             = var.isolation_segments == "1" && var.dualstack ? 1 : 0
  load_balancer_arn = aws_lb.iso_router_nlb[0].arn
  port              = "443"
  protocol          = "TLS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.lb_cert.arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.iso_router_nlb_http[0].arn
  }
}

resource "aws_lb_listener" "iso_router_nlb_4443" {
  count             = var.isolation_segments == "1" && var.dualstack ? 1 : 0
  load_balancer_arn = aws_lb.iso_router_nlb[0].arn
  port              = "4443"
  protocol          = "TLS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.lb_cert.arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.iso_router_nlb_http[0].arn
  }
}

resource "aws_lb_target_group" "iso_router_lb_4443" {
  count    = var.isolation_segments
  name     = "${var.short_env_id}-isotg-4443"
  port     = 4443
  protocol = "TCP"
  vpc_id   = local.vpc_id

  health_check {
    protocol = "TCP"
  }

  tags = {
    Name = "${var.env_id}"
  }
}

resource "aws_security_group" "iso_security_group" {
  count = var.isolation_segments

  name   = "${var.env_id}-iso-sg"
  vpc_id = local.vpc_id

  description = "Private isolation segment"

  tags = {
    Name = "${var.env_id}-iso-security-group"
  }
}

resource "aws_security_group" "iso_shared_security_group" {
  count = var.isolation_segments

  name   = "${var.env_id}-iso-shared-sg"
  vpc_id = local.vpc_id

  description = "Shared isolation segments"

  tags = {
    Name = "${var.env_id}-iso-shared-security-group"
  }
}

resource "aws_security_group_rule" "isolation_segments_to_bosh_rule" {
  count = var.isolation_segments * length(var.iso_to_bosh_ports)

  description = "TCP traffic from iso-sg to bosh"

  security_group_id        = aws_security_group.bosh_security_group[count.index].id
  type                     = "ingress"
  protocol                 = "tcp"
  to_port                  = element(var.iso_to_bosh_ports, count.index)
  from_port                = element(var.iso_to_bosh_ports, count.index)
  source_security_group_id = aws_security_group.iso_security_group[count.index].id
}

resource "aws_security_group_rule" "isolation_segments_to_shared_tcp_rule" {
  count = var.isolation_segments * length(var.iso_to_shared_tcp_ports)

  description = "TCP traffic from iso-sg to iso-shared-sg"

  security_group_id        = aws_security_group.iso_shared_security_group[count.index].id
  type                     = "ingress"
  protocol                 = "tcp"
  to_port                  = element(var.iso_to_shared_tcp_ports, count.index)
  from_port                = element(var.iso_to_shared_tcp_ports, count.index)
  source_security_group_id = aws_security_group.iso_security_group[count.index].id
}

resource "aws_security_group_rule" "isolation_segments_to_shared_udp_rule" {
  count = var.isolation_segments * length(var.iso_to_shared_udp_ports)

  description = "UDP traffic from iso-sg to iso-shared-sg"

  security_group_id        = aws_security_group.iso_shared_security_group[count.index].id
  type                     = "ingress"
  protocol                 = "udp"
  to_port                  = element(var.iso_to_shared_udp_ports, count.index)
  from_port                = element(var.iso_to_shared_udp_ports, count.index)
  source_security_group_id = aws_security_group.iso_security_group[count.index].id
}

resource "aws_security_group_rule" "isolation_segments_to_bosh_all_traffic_rule" {
  count = var.isolation_segments

  description = "ALL traffic from iso-sg to bosh"

  depends_on               = [aws_security_group.bosh_security_group]
  security_group_id        = aws_security_group.bosh_security_group[count.index].id
  type                     = "ingress"
  protocol                 = "-1"
  from_port                = 0
  to_port                  = 0
  source_security_group_id = aws_security_group.iso_security_group[count.index].id
}

resource "aws_security_group_rule" "shared_diego_bbs_to_isolated_cells_rule" {
  count = var.isolation_segments

  description = "TCP traffic from shared diego bbs to iso-sg"

  depends_on               = [aws_security_group.iso_security_group]
  security_group_id        = aws_security_group.iso_security_group[count.index].id
  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 1801
  to_port                  = 1801
  source_security_group_id = aws_security_group.iso_shared_security_group[count.index].id
}

resource "aws_security_group_rule" "nat_to_isolated_cells_rule" {
  count = var.isolation_segments

  description = "ALL traffic from nat-sg to iso-sg"

  security_group_id        = aws_security_group.nat_security_group[count.index].id
  type                     = "ingress"
  protocol                 = "-1"
  from_port                = 0
  to_port                  = 0
  source_security_group_id = aws_security_group.iso_security_group[count.index].id
}

output "cf_iso_router_lb_name" {
  value = var.dualstack ? one(aws_lb.iso_router_nlb[*].name) : one(aws_elb.iso_router_lb[*].name)
}

output "iso_security_group_id" {
  value = one(aws_security_group.iso_security_group[*].id)
}

output "iso_az_subnet_id_mapping" {
  value = zipmap("${aws_subnet.iso_subnets.*.availability_zone}", "${aws_subnet.iso_subnets.*.id}")
}

output "iso_az_subnet_cidr_mapping" {
  value = zipmap("${aws_subnet.iso_subnets.*.availability_zone}", "${aws_subnet.iso_subnets.*.cidr_block}")
}

output "iso_az_subnet_ipv6_cidr_mapping" {
  value = var.dualstack ? "${zipmap("${aws_subnet.iso_subnets.*.availability_zone}", "${aws_subnet.iso_subnets.*.cidr_block}")}" : null
}

output "iso_shared_security_group_id" {
  value = one(aws_security_group.iso_shared_security_group[*].id)
}
