resource "aws_instance" "nat" {
  count = 0
}

resource "aws_eip" "nat_eip" {
  instance = ""
}

resource "aws_nat_gateway" "nat" {
  allocation_id = "${aws_eip.nat_eip.id}"
  subnet_id     = "${aws_subnet.bosh_subnet.id}"
  depends_on    = ["aws_internet_gateway.ig"]

  tags {
    Name  = "${var.env_id}-nat"
    EnvID = "${var.env_id}"
  }
}

resource "aws_route" "internal_route_table" {
  instance_id    = ""
  nat_gateway_id = "${aws_nat_gateway.nat.id}"
}
