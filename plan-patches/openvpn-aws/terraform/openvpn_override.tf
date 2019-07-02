resource "aws_subnet" "openvpn_subnet" {
  vpc_id            = "${local.vpc_id}"
  cidr_block        = "${cidrsubnet(var.vpc_cidr, 8, 144)}"
  availability_zone = "${element(var.availability_zones, 0)}"

  tags {
    Name = "${var.env_id}-openvpn-subnet"
  }

  lifecycle {
    ignore_changes = ["cidr_block", "availability_zone"]
  }
}

resource "aws_route_table" "openvpn_route_table" {
  vpc_id = "${local.vpc_id}"
}

resource "aws_route" "openvpn_route" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.ig.id}"
  route_table_id         = "${aws_route_table.openvpn_route_table.id}"
}

resource "aws_route_table_association" "openvpn_route_table_association" {
  subnet_id      = "${aws_subnet.openvpn_subnet.id}"
  route_table_id = "${aws_route_table.openvpn_route_table.id}"
}

resource "aws_security_group" "openvpn_security_group" {
  name        = "${var.env_id}-openvpn-security-group"
  description = "OpenVPN"
  vpc_id      = "${local.vpc_id}"

  tags {
    Name = "${var.env_id}-openvpn-security-group"
  }

  lifecycle {
    ignore_changes = ["name"]
  }
}

resource "aws_security_group_rule" "openvpn_security_group_rule_tcp" {
  security_group_id = "${aws_security_group.openvpn_security_group.id}"
  type              = "ingress"
  protocol          = "tcp"
  from_port         = 0
  to_port           = 65535
  self              = true
}

resource "aws_security_group_rule" "openvpn_security_group_rule_udp" {
  security_group_id = "${aws_security_group.openvpn_security_group.id}"
  type              = "ingress"
  protocol          = "udp"
  from_port         = 0
  to_port           = 65535
  self              = true
}

resource "aws_security_group_rule" "openvpn_security_group_rule_icmp" {
  security_group_id = "${aws_security_group.openvpn_security_group.id}"
  type              = "ingress"
  protocol          = "icmp"
  from_port         = -1
  to_port           = -1
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "openvpn_security_group_rule_openvpn" {
  security_group_id = "${aws_security_group.openvpn_security_group.id}"
  type              = "ingress"
  protocol          = "TCP"
  from_port         = 1194
  to_port           = 1194
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "openvpn_security_group_rule_allow_internet" {
  security_group_id = "${aws_security_group.openvpn_security_group.id}"
  type              = "egress"
  protocol          = "-1"
  from_port         = 0
  to_port           = 0
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "openvpn_security_group_rule_ssh" {
  security_group_id        = "${aws_security_group.openvpn_security_group.id}"
  type                     = "ingress"
  protocol                 = "TCP"
  from_port                = 22
  to_port                  = 22
  source_security_group_id = "${aws_security_group.jumpbox.id}"
}

resource "aws_security_group_rule" "bosh_openvpn_security_group_rule_tcp" {
  security_group_id        = "${aws_security_group.openvpn_security_group.id}"
  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = "${aws_security_group.bosh_security_group.id}"
}

resource "aws_security_group_rule" "bosh_openvpn_security_group_rule_udp" {
  security_group_id        = "${aws_security_group.openvpn_security_group.id}"
  type                     = "ingress"
  protocol                 = "udp"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = "${aws_security_group.bosh_security_group.id}"
}

resource "aws_security_group_rule" "openvpn_bosh_security_group_rule_tcp" {
  security_group_id        = "${aws_security_group.bosh_security_group.id}"
  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = "${aws_security_group.openvpn_security_group.id}"
}

resource "aws_security_group_rule" "openvpn_bosh_security_group_rule_udp" {
  security_group_id        = "${aws_security_group.bosh_security_group.id}"
  type                     = "ingress"
  protocol                 = "udp"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = "${aws_security_group.openvpn_security_group.id}"
}

resource "aws_eip" "openvpn_eip" {
  vpc = true
}

output "openvpn_ip" {
  value = "${aws_eip.openvpn_eip.public_ip}"
}

output "openvpn_security_group" {
  value = "${aws_security_group.openvpn_security_group.id}"
}

output "openvpn_subnet_cidr" {
  value = "${aws_subnet.openvpn_subnet.cidr_block}"
}

output "openvpn_subnet_id" {
  value = "${aws_subnet.openvpn_subnet.id}"
}
