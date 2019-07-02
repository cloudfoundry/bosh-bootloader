resource "aws_subnet" "nat_subnet" {
  count             = "${length(var.availability_zones)}"

  vpc_id            = "${local.vpc_id}"
  cidr_block        = "${cidrsubnet(var.vpc_cidr, 12, count.index + 16)}"
  availability_zone = "${element(var.availability_zones, count.index)}"

  tags {
    Name = "${var.env_id}-nat-subnet"
  }
}

resource "aws_route_table" "nat_route_table" {
  vpc_id = "${local.vpc_id}"

  tags {
    Name = "${var.env_id}-nat-rtb"
  }
}

resource "aws_route" "nat_route" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.ig.id}"
  route_table_id         = "${aws_route_table.nat_route_table.id}"
}

resource "aws_route_table_association" "route_nat_subnets" {
  count          = "${length(var.availability_zones)}"

  subnet_id      = "${element(aws_subnet.nat_subnet.*.id, count.index)}"
  route_table_id = "${element(aws_route_table.nat_route_table.*.id, count.index)}"
}

resource "aws_instance" "nat" {
  count                  = "${length(var.availability_zones)}"

  private_ip             = "${cidrhost(element(aws_subnet.nat_subnet.*.cidr_block, count.index), 7)}"
  instance_type          = "t2.medium"
  subnet_id              = "${element(aws_subnet.nat_subnet.*.id, count.index)}"
  source_dest_check      = false
  ami                    = "${lookup(var.nat_ami_map, var.region)}"
  vpc_security_group_ids = ["${aws_security_group.nat_security_group.id}"]

  tags {
    Name  = "${var.env_id}-nat"
    EnvID = "${var.env_id}"
  }
}

resource "aws_eip" "nat_eip" {
  count      = "${length(var.availability_zones)}"

  depends_on = ["aws_internet_gateway.ig"]
  instance   = "${element(aws_instance.nat.*.id, count.index)}"
  vpc        = true

  tags {
    Name  = "${var.env_id}-nat_eip-${count.index}"
    EnvID = "${var.env_id}"
  }
}

resource "aws_route_table" "internal_route_table" {
  count = "${length(var.availability_zones)}"

  vpc_id = "${local.vpc_id}"

  tags {
    Name = "${var.env_id}-private-rtb-${count.index}"
  }
}

resource "aws_route" "internal_route_table" {
  count = "${length(var.availability_zones)}"

  destination_cidr_block = "0.0.0.0/0"
  instance_id            = "${element(aws_instance.nat.*.id, count.index)}"
  route_table_id         = "${element(aws_route_table.internal_route_table.*.id, count.index)}"
}

resource "aws_route_table_association" "route_internal_subnets" {
  count          = "${length(var.availability_zones)}"

  subnet_id      = "${element(aws_subnet.internal_subnets.*.id, count.index)}"
  route_table_id = "${element(aws_route_table.internal_route_table.*.id, count.index)}"
}

output "nat_eip" {
  value = "${aws_eip.nat_eip.*.public_ip}"
}
