resource "aws_subnet" "lb_subnets" {
  count             = "${length(var.availability_zones)}"
  vpc_id            = "${local.vpc_id}"
  cidr_block        = "${cidrsubnet(var.vpc_cidr, 8, count.index+2)}"
  ipv6_cidr_block   = var.dualstack ? "${cidrsubnet(aws_vpc.vpc[0].ipv6_cidr_block, 8, count.index + 1 + length(var.availability_zones))}" : null
  availability_zone = "${element(var.availability_zones, count.index)}"
  assign_ipv6_address_on_creation = var.dualstack
  enable_dns64 = var.dualstack

  tags = {
    Name = "${var.env_id}-lb-subnet${count.index}"
  }

  lifecycle {
    ignore_changes = [cidr_block, availability_zone]
  }
}

resource "aws_route_table" "lb_route_table" {
  vpc_id = "${local.vpc_id}"
}

resource "aws_route" "lb_route_table" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.ig.id}"
  route_table_id         = "${aws_route_table.lb_route_table.id}"
}

resource "aws_route" "lb_route_table_ipv6" {
  count = var.dualstack ? 1 : 0
  route_table_id         = "${aws_route_table.lb_route_table.id}"
  destination_ipv6_cidr_block = "::/0"
  egress_only_gateway_id = aws_egress_only_internet_gateway.egress_ipv6[0].id
}

resource "aws_route_table_association" "route_lb_subnets" {
  count          = "${length(var.availability_zones)}"
  subnet_id      = "${element(aws_subnet.lb_subnets.*.id, count.index)}"
  route_table_id = "${aws_route_table.lb_route_table.id}"
}

output "lb_subnet_ids" {
  value = ["${aws_subnet.lb_subnets.*.id}"]
}

output "lb_subnet_availability_zones" {
  value = ["${aws_subnet.lb_subnets.*.availability_zone}"]
}

output "lb_subnet_cidrs" {
  value = ["${aws_subnet.lb_subnets.*.cidr_block}"]
}

output "lb_subnet_ipv6_cidrs" {
  value = var.dualstack ? aws_subnet.lb_subnets.*.ipv6_cidr_block : null
}
