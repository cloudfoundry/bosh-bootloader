provider "aws" {
  access_key = "${var.access_key}"
  secret_key = "${var.secret_key}"
  region     = "${var.region}"

  version = "~> 1.60"
}

provider "tls" {
  version = "~> 1.2"
}

variable "access_key" {
  type = "string"
}

variable "secret_key" {
  type = "string"
}

variable "region" {
  type = "string"
}

variable "bosh_inbound_cidr" {
  default = "0.0.0.0/0"
}

variable "availability_zones" {
  type = "list"
}

variable "env_id" {
  type = "string"
}

variable "short_env_id" {
  type = "string"
}

variable "vpc_cidr" {
  type    = "string"
  default = "10.0.0.0/16"
}

resource "aws_eip" "jumpbox_eip" {
  depends_on = ["aws_internet_gateway.ig"]
  vpc        = true
}

resource "tls_private_key" "bosh_vms" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "bosh_vms" {
  key_name   = "${var.env_id}_bosh_vms"
  public_key = "${tls_private_key.bosh_vms.public_key_openssh}"
}

resource "aws_security_group" "nat_security_group" {
  name        = "${var.env_id}-nat-security-group"
  description = "NAT"
  vpc_id      = "${local.vpc_id}"

  tags {
    Name = "${var.env_id}-nat-security-group"
  }

  lifecycle {
    ignore_changes = ["name"]
  }
}

resource "aws_security_group_rule" "nat_to_internet_rule" {
  security_group_id = "${aws_security_group.nat_security_group.id}"

  type        = "egress"
  from_port   = 0
  to_port     = 0
  protocol    = "-1"
  cidr_blocks = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "nat_icmp_rule" {
  security_group_id = "${aws_security_group.nat_security_group.id}"

  type        = "ingress"
  protocol    = "icmp"
  from_port   = -1
  to_port     = -1
  cidr_blocks = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "nat_tcp_rule" {
  security_group_id = "${aws_security_group.nat_security_group.id}"

  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = "${aws_security_group.internal_security_group.id}"
}

resource "aws_security_group_rule" "nat_udp_rule" {
  security_group_id = "${aws_security_group.nat_security_group.id}"

  type                     = "ingress"
  protocol                 = "udp"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = "${aws_security_group.internal_security_group.id}"
}

resource "aws_nat_gateway" "nat" {
  subnet_id     = "${aws_subnet.bosh_subnet.id}"
  allocation_id = "${aws_eip.nat_eip.id}"

  tags {
    Name  = "${var.env_id}-nat"
    EnvID = "${var.env_id}"
  }
}

resource "aws_eip" "nat_eip" {
  vpc = true

  tags {
    Name  = "${var.env_id}-nat"
    EnvID = "${var.env_id}"
  }
}

resource "aws_default_security_group" "default_security_group" {
  vpc_id = "${local.vpc_id}"
}

resource "aws_security_group" "internal_security_group" {
  name        = "${var.env_id}-internal-security-group"
  description = "Internal"
  vpc_id      = "${local.vpc_id}"

  tags {
    Name = "${var.env_id}-internal-security-group"
  }

  lifecycle {
    ignore_changes = ["name"]
  }
}

resource "aws_security_group_rule" "internal_security_group_rule_tcp" {
  security_group_id = "${aws_security_group.internal_security_group.id}"
  type              = "ingress"
  protocol          = "tcp"
  from_port         = 0
  to_port           = 65535
  self              = true
}

resource "aws_security_group_rule" "internal_security_group_rule_udp" {
  security_group_id = "${aws_security_group.internal_security_group.id}"
  type              = "ingress"
  protocol          = "udp"
  from_port         = 0
  to_port           = 65535
  self              = true
}

resource "aws_security_group_rule" "internal_security_group_rule_icmp" {
  security_group_id = "${aws_security_group.internal_security_group.id}"
  type              = "ingress"
  protocol          = "icmp"
  from_port         = -1
  to_port           = -1
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "internal_security_group_rule_allow_internet" {
  security_group_id = "${aws_security_group.internal_security_group.id}"
  type              = "egress"
  protocol          = "-1"
  from_port         = 0
  to_port           = 0
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "internal_security_group_rule_ssh" {
  security_group_id        = "${aws_security_group.internal_security_group.id}"
  type                     = "ingress"
  protocol                 = "TCP"
  from_port                = 22
  to_port                  = 22
  source_security_group_id = "${aws_security_group.jumpbox.id}"
}

resource "aws_security_group" "bosh_security_group" {
  name        = "${var.env_id}-bosh-security-group"
  description = "BOSH Director"
  vpc_id      = "${local.vpc_id}"

  tags {
    Name = "${var.env_id}-bosh-security-group"
  }

  lifecycle {
    ignore_changes = ["name", "description"]
  }
}

resource "aws_security_group_rule" "bosh_security_group_rule_tcp_ssh" {
  security_group_id        = "${aws_security_group.bosh_security_group.id}"
  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 22
  to_port                  = 22
  source_security_group_id = "${aws_security_group.jumpbox.id}"
}

resource "aws_security_group_rule" "bosh_security_group_rule_tcp_bosh_agent" {
  security_group_id        = "${aws_security_group.bosh_security_group.id}"
  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 6868
  to_port                  = 6868
  source_security_group_id = "${aws_security_group.jumpbox.id}"
}

resource "aws_security_group_rule" "bosh_security_group_rule_uaa" {
  security_group_id        = "${aws_security_group.bosh_security_group.id}"
  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 8443
  to_port                  = 8443
  source_security_group_id = "${aws_security_group.jumpbox.id}"
}

resource "aws_security_group_rule" "bosh_security_group_rule_credhub" {
  security_group_id        = "${aws_security_group.bosh_security_group.id}"
  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 8844
  to_port                  = 8844
  source_security_group_id = "${aws_security_group.jumpbox.id}"
}

resource "aws_security_group_rule" "bosh_security_group_rule_tcp_director_api" {
  security_group_id        = "${aws_security_group.bosh_security_group.id}"
  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 25555
  to_port                  = 25555
  source_security_group_id = "${aws_security_group.jumpbox.id}"
}

resource "aws_security_group_rule" "bosh_security_group_rule_tcp" {
  security_group_id        = "${aws_security_group.bosh_security_group.id}"
  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = "${aws_security_group.internal_security_group.id}"
}

resource "aws_security_group_rule" "bosh_security_group_rule_udp" {
  security_group_id        = "${aws_security_group.bosh_security_group.id}"
  type                     = "ingress"
  protocol                 = "udp"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = "${aws_security_group.internal_security_group.id}"
}

resource "aws_security_group_rule" "bosh_security_group_rule_allow_internet" {
  security_group_id = "${aws_security_group.bosh_security_group.id}"
  type              = "egress"
  protocol          = "-1"
  from_port         = 0
  to_port           = 0
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group" "jumpbox" {
  name        = "${var.env_id}-jumpbox-security-group"
  description = "Jumpbox"
  vpc_id      = "${local.vpc_id}"

  tags {
    Name = "${var.env_id}-jumpbox-security-group"
  }

  lifecycle {
    ignore_changes = ["name", "description"]
  }
}

resource "aws_security_group_rule" "jumpbox_ssh" {
  security_group_id = "${aws_security_group.jumpbox.id}"
  type              = "ingress"
  protocol          = "tcp"
  from_port         = 22
  to_port           = 22
  cidr_blocks       = ["${var.bosh_inbound_cidr}"]
}

resource "aws_security_group_rule" "jumpbox_rdp" {
  security_group_id = "${aws_security_group.jumpbox.id}"
  type              = "ingress"
  protocol          = "tcp"
  from_port         = 3389
  to_port           = 3389
  cidr_blocks       = ["${var.bosh_inbound_cidr}"]
}

resource "aws_security_group_rule" "jumpbox_agent" {
  security_group_id = "${aws_security_group.jumpbox.id}"
  type              = "ingress"
  protocol          = "tcp"
  from_port         = 6868
  to_port           = 6868
  cidr_blocks       = ["${var.bosh_inbound_cidr}"]
}

resource "aws_security_group_rule" "jumpbox_director" {
  security_group_id = "${aws_security_group.jumpbox.id}"
  type              = "ingress"
  protocol          = "tcp"
  from_port         = 25555
  to_port           = 25555
  cidr_blocks       = ["${var.bosh_inbound_cidr}"]
}

resource "aws_security_group_rule" "jumpbox_egress" {
  security_group_id = "${aws_security_group.jumpbox.id}"
  type              = "egress"
  protocol          = "-1"
  from_port         = 0
  to_port           = 0
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "bosh_internal_security_rule_tcp" {
  security_group_id        = "${aws_security_group.internal_security_group.id}"
  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = "${aws_security_group.bosh_security_group.id}"
}

resource "aws_security_group_rule" "bosh_internal_security_rule_udp" {
  security_group_id        = "${aws_security_group.internal_security_group.id}"
  type                     = "ingress"
  protocol                 = "udp"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = "${aws_security_group.bosh_security_group.id}"
}

resource "aws_subnet" "bosh_subnet" {
  vpc_id     = "${local.vpc_id}"
  cidr_block = "${cidrsubnet(var.vpc_cidr, 8, 0)}"

  tags {
    Name = "${var.env_id}-bosh-subnet"
  }
}

resource "aws_route_table" "bosh_route_table" {
  vpc_id = "${local.vpc_id}"
}

resource "aws_route" "bosh_route_table" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.ig.id}"
  route_table_id         = "${aws_route_table.bosh_route_table.id}"
}

resource "aws_route_table_association" "route_bosh_subnets" {
  subnet_id      = "${aws_subnet.bosh_subnet.id}"
  route_table_id = "${aws_route_table.bosh_route_table.id}"
}

resource "aws_subnet" "internal_subnets" {
  count             = "${length(var.availability_zones)}"
  vpc_id            = "${local.vpc_id}"
  cidr_block        = "${cidrsubnet(var.vpc_cidr, 4, count.index+1)}"
  availability_zone = "${element(var.availability_zones, count.index)}"

  tags {
    Name = "${var.env_id}-internal-subnet${count.index}"
  }

  lifecycle {
    ignore_changes = ["cidr_block", "availability_zone"]
  }
}

resource "aws_route_table" "nated_route_table" {
  vpc_id = "${local.vpc_id}"

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = "${aws_nat_gateway.nat.id}"
  }
}

resource "aws_route_table_association" "route_internal_subnets" {
  count          = "${length(var.availability_zones)}"
  subnet_id      = "${element(aws_subnet.internal_subnets.*.id, count.index)}"
  route_table_id = "${aws_route_table.nated_route_table.id}"
}

resource "aws_internet_gateway" "ig" {
  vpc_id = "${local.vpc_id}"
}

locals {
  director_name        = "bosh-${var.env_id}"
  internal_cidr        = "${aws_subnet.bosh_subnet.cidr_block}"
  internal_gw          = "${cidrhost(local.internal_cidr, 1)}"
  jumpbox_internal_ip  = "${cidrhost(local.internal_cidr, 5)}"
  director_internal_ip = "${cidrhost(local.internal_cidr, 6)}"
}

resource "aws_kms_key" "kms_key" {
  enable_key_rotation = true
}

output "default_key_name" {
  value = "${aws_key_pair.bosh_vms.key_name}"
}

output "private_key" {
  value     = "${tls_private_key.bosh_vms.private_key_pem}"
  sensitive = true
}

output "external_ip" {
  value = "${aws_eip.jumpbox_eip.public_ip}"
}

output "jumpbox_url" {
  value = "${aws_eip.jumpbox_eip.public_ip}:22"
}

output "director_address" {
  value = "https://${aws_eip.jumpbox_eip.public_ip}:25555"
}

output "nat_eip" {
  value = "${aws_eip.nat_eip.public_ip}"
}

output "internal_security_group" {
  value = "${aws_security_group.internal_security_group.id}"
}

output "bosh_security_group" {
  value = "${aws_security_group.bosh_security_group.id}"
}

output "jumpbox_security_group" {
  value = "${aws_security_group.jumpbox.id}"
}

output "jumpbox__default_security_groups" {
  value = ["${aws_security_group.jumpbox.id}"]
}

output "director__default_security_groups" {
  value = ["${aws_security_group.bosh_security_group.id}"]
}

output "subnet_id" {
  value = "${aws_subnet.bosh_subnet.id}"
}

output "az" {
  value = "${aws_subnet.bosh_subnet.availability_zone}"
}

output "vpc_id" {
  value = "${local.vpc_id}"
}

output "region" {
  value = "${var.region}"
}

output "kms_key_arn" {
  value = "${aws_kms_key.kms_key.arn}"
}

output "internal_az_subnet_id_mapping" {
  value = "${
	  zipmap("${aws_subnet.internal_subnets.*.availability_zone}", "${aws_subnet.internal_subnets.*.id}")
	}"
}

output "internal_az_subnet_cidr_mapping" {
  value = "${
	  zipmap("${aws_subnet.internal_subnets.*.availability_zone}", "${aws_subnet.internal_subnets.*.cidr_block}")
	}"
}

output "director_name" {
  value = "${local.director_name}"
}

output "internal_cidr" {
  value = "${local.internal_cidr}"
}

output "internal_gw" {
  value = "${local.internal_gw}"
}

output "jumpbox__internal_ip" {
  value = "${local.jumpbox_internal_ip}"
}

output "director__internal_ip" {
  value = "${local.director_internal_ip}"
}
