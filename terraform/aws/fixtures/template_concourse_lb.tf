resource "aws_eip" "bosh_eip" {
  depends_on = ["aws_internet_gateway.ig"]
  vpc      = true
}

output "bosh_eip" {
  value = "${aws_eip.bosh_eip.public_ip}"
}

output "bosh_url" {
  value = "https://${aws_eip.bosh_eip.public_ip}:25555"
}

resource "aws_iam_user" "bosh" {
  name = "${var.env_id}_bosh_user"
}

resource "aws_iam_user_policy" "bosh" {
  name  = "${var.env_id}_bosh_user_policy"
  user = "${aws_iam_user.bosh.name}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:AssociateAddress",
        "ec2:AttachVolume",
        "ec2:CreateVolume",
        "ec2:DeleteSnapshot",
        "ec2:DeleteVolume",
        "ec2:DescribeAddresses",
        "ec2:DescribeImages",
        "ec2:DescribeInstances",
        "ec2:DescribeRegions",
        "ec2:DescribeSecurityGroups",
        "ec2:DescribeSnapshots",
        "ec2:DescribeSubnets",
        "ec2:DescribeVolumes",
        "ec2:DetachVolume",
        "ec2:CreateSnapshot",
        "ec2:CreateTags",
        "ec2:RunInstances",
        "ec2:TerminateInstances",
        "ec2:RegisterImage",
        "ec2:DeregisterImage"
      ],
      "Effect": "Allow",
      "Resource": "*"
    },
    {
      "Action": [
        "elasticloadbalancing:*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_access_key" "bosh" {
  user = "${aws_iam_user.bosh.name}"
}

output "bosh_user_access_key" {
  value = "${aws_iam_access_key.bosh.id}"
}

output "bosh_user_secret_access_key" {
  value = "${aws_iam_access_key.bosh.secret}"
}

variable "nat_ami_map" {
  type = "map"

  default = {
    us-east-1      ="ami-68115b02"
    us-west-1      ="ami-ef1a718f"
    us-west-2      ="ami-77a4b816"
    eu-west-1      ="ami-c0993ab3"
    eu-central-1   ="ami-0b322e67"
    ap-southeast-1 ="ami-e2fc3f81"
    ap-southeast-2 ="ami-e3217a80"
    ap-northeast-1 ="ami-f885ae96"
    ap-northeast-2 ="ami-4118d72f"
    sa-east-1      ="ami-8631b5ea"
  }
}

resource "aws_security_group" "nat_security_group" {
  description = "NAT"
  vpc_id      = "${aws_vpc.vpc.id}"

  ingress {
    protocol    = "tcp"
    from_port   = 0
    to_port     = 65535
    security_groups = ["${aws_security_group.internal_security_group.id}"]
  }

  ingress {
    protocol    = "udp"
    from_port   = 0
    to_port     = 65535
    security_groups = ["${aws_security_group.internal_security_group.id}"]
  }

  ingress {
    protocol    = "icmp"
    from_port   = -1
    to_port     = -1
    security_groups = ["${aws_security_group.internal_security_group.id}"]
  }

  egress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    Name = "${var.env_id}-nat-security-group"
  }
}

variable "nat_ssh_key_pair_name" {}

resource "aws_instance" "nat" {
  private_ip             = "10.0.0.7"
  instance_type          = "t2.medium"
  subnet_id              = "${aws_subnet.bosh_subnet.id}"
  source_dest_check      = false
  ami                    = "${lookup(var.nat_ami_map, var.region)}"
  key_name               = "${var.nat_ssh_key_pair_name}"
  vpc_security_group_ids = ["${aws_security_group.nat_security_group.id}"]

  tags {
    Name = "${var.env_id}-nat",
    EnvID = "${var.env_id}"
  }
}

resource "aws_eip" "nat_eip" {
  depends_on = ["aws_internet_gateway.ig"]
  instance = "${aws_instance.nat.id}"
  vpc      = true
}

output "nat_eip" {
  value = "${aws_eip.nat_eip.public_ip}"
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

provider "aws" {
  access_key = "${var.access_key}"
  secret_key = "${var.secret_key}"
  region     = "${var.region}"
}

resource "aws_security_group" "internal_security_group" {
  description = "Internal"
  vpc_id      = "${aws_vpc.vpc.id}"

  tags {
    Name = "${var.env_id}-internal-security-group"
  }
}

resource "aws_security_group_rule" "internal_security_group_rule_tcp" {
  security_group_id        = "${aws_security_group.internal_security_group.id}"
  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 0
  to_port                  = 65535
  self                     = true
}

resource "aws_security_group_rule" "internal_security_group_rule_udp" {
  security_group_id        = "${aws_security_group.internal_security_group.id}"
  type                     = "ingress"
  protocol                 = "udp"
  from_port                = 0
  to_port                  = 65535
  self                     = true
}

resource "aws_security_group_rule" "internal_security_group_rule_icmp" {
  security_group_id        = "${aws_security_group.internal_security_group.id}"
  type                     = "ingress"
  protocol                 = "icmp"
  from_port                = -1
  to_port                  = -1
  cidr_blocks              = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "internal_security_group_rule_allow_internet" {
  security_group_id        = "${aws_security_group.internal_security_group.id}"
  type                     = "egress"
  protocol                 = "-1"
  from_port                = 0
  to_port                  = 0
  cidr_blocks              = ["0.0.0.0/0"]
}

output "internal_security_group" {
  value="${aws_security_group.internal_security_group.id}"
}

variable "bosh_inbound_cidr" {
  default = "0.0.0.0/0"
}

resource "aws_security_group" "bosh_security_group" {
  description = "Bosh"
  vpc_id      = "${aws_vpc.vpc.id}"

  tags {
    Name = "${var.env_id}-bosh-security-group"
  }
}

resource "aws_security_group_rule" "bosh_security_group_rule_tcp_ssh" {
  security_group_id        = "${aws_security_group.bosh_security_group.id}"
  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 22
  to_port                  = 22
  cidr_blocks              = ["${var.bosh_inbound_cidr}"]
}

resource "aws_security_group_rule" "bosh_security_group_rule_tcp_bosh_agent" {
  security_group_id        = "${aws_security_group.bosh_security_group.id}"
  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 6868
  to_port                  = 6868
  cidr_blocks              = ["${var.bosh_inbound_cidr}"]
}

resource "aws_security_group_rule" "bosh_security_group_rule_tcp_director_api" {
  security_group_id        = "${aws_security_group.bosh_security_group.id}"
  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 25555
  to_port                  = 25555
  cidr_blocks              = ["${var.bosh_inbound_cidr}"]
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
  security_group_id        = "${aws_security_group.bosh_security_group.id}"
  type                     = "egress"
  protocol                 = "-1"
  from_port                = 0
  to_port                  = 0
  cidr_blocks              = ["0.0.0.0/0"]
}

output "bosh_security_group" {
  value="${aws_security_group.bosh_security_group.id}"
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

variable "bosh_subnet_cidr" {
  type    = "string"
  default = "10.0.0.0/24"
}

variable "bosh_availability_zone" {
  type = "string"
}

resource "aws_subnet" "bosh_subnet" {
  vpc_id            = "${aws_vpc.vpc.id}"
  cidr_block        = "${var.bosh_subnet_cidr}"
  availability_zone = "${var.bosh_availability_zone}"

  tags {
    Name = "${var.env_id}-bosh-subnet"
  }
}

resource "aws_route_table" "bosh_route_table" {
  vpc_id = "${aws_vpc.vpc.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.ig.id}"
  }
}

resource "aws_route_table_association" "route_bosh_subnets" {
  subnet_id      = "${aws_subnet.bosh_subnet.id}"
  route_table_id = "${aws_route_table.bosh_route_table.id}"
}

output "bosh_subnet_id" {
  value = "${aws_subnet.bosh_subnet.id}"
}

output "bosh_subnet_availability_zone" {
  value = "${aws_subnet.bosh_subnet.availability_zone}"
}

variable "availability_zones" {
  type = "list"
}

resource "aws_subnet" "internal_subnets" {
  count             = "${length(var.availability_zones)}"
  vpc_id            = "${aws_vpc.vpc.id}"
  cidr_block        = "${cidrsubnet("10.0.0.0/16", 4, count.index+1)}"
  availability_zone = "${element(var.availability_zones, count.index)}"

  tags {
    Name = "${var.env_id}-internal-subnet${count.index}"
  }
}

resource "aws_route_table" "internal_route_table" {
  vpc_id = "${aws_vpc.vpc.id}"

  route {
    cidr_block = "0.0.0.0/0"
    instance_id = "${aws_instance.nat.id}"
  }
}

resource "aws_route_table_association" "route_internal_subnets" {
  count          = "${length(var.availability_zones)}"
  subnet_id      = "${element(aws_subnet.internal_subnets.*.id, count.index)}"
  route_table_id = "${aws_route_table.internal_route_table.id}"
}

output "internal_subnet_ids" {
  value = ["${aws_subnet.internal_subnets.*.id}"]
}

output "internal_subnet_availability_zones" {
  value = ["${aws_subnet.internal_subnets.*.availability_zone}"]
}

output "internal_subnet_cidrs" {
  value = ["${aws_subnet.internal_subnets.*.cidr_block}"]
}

variable "env_id" {
  type = "string"
}

variable "short_env_id" {
  type = "string"
}

variable "vpc_cidr" {
  type = "string"
  default = "10.0.0.0/16"
}

resource "aws_vpc" "vpc" {
  cidr_block           = "${var.vpc_cidr}"
  instance_tenancy     = "default"
  enable_dns_hostnames = true

  tags {
    Name = "${var.env_id}-vpc"
  }
}

resource "aws_internet_gateway" "ig" {
  vpc_id = "${aws_vpc.vpc.id}"
}

output "vpc_id" {
  value = "${aws_vpc.vpc.id}"
}

resource "aws_subnet" "lb_subnets" {
  count             = "${length(var.availability_zones)}"
  vpc_id            = "${aws_vpc.vpc.id}"
  cidr_block        = "${cidrsubnet("10.0.0.0/20", 4, count.index+2)}"
  availability_zone = "${element(var.availability_zones, count.index)}"

  tags {
    Name = "${var.env_id}-lb-subnet${count.index}"
  }
}

resource "aws_route_table" "lb_route_table" {
  vpc_id = "${aws_vpc.vpc.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.ig.id}"
  }
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

variable "ssl_certificate" {
  type = "string"
}

variable "ssl_certificate_chain" {
  type = "string"
}

variable "ssl_certificate_private_key" {
  type = "string"
}

resource "aws_iam_server_certificate" "lb_cert" {
  name_prefix       = "${var.short_env_id}-"

  certificate_body  = "${var.ssl_certificate}"
  certificate_chain = "${var.ssl_certificate_chain}"
  private_key       = "${var.ssl_certificate_private_key}"

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_security_group" "concourse_lb_security_group" {
  description = "Concourse"
  vpc_id      = "${aws_vpc.vpc.id}"

  ingress {
    cidr_blocks = ["0.0.0.0/0"]
    protocol    = "tcp"
    from_port   = 80
    to_port     = 80
  }

  ingress {
    cidr_blocks = ["0.0.0.0/0"]
    protocol    = "tcp"
    from_port   = 2222
    to_port     = 2222
  }

  ingress {
    cidr_blocks = ["0.0.0.0/0"]
    protocol    = "tcp"
    from_port   = 443
    to_port     = 443
  }

  egress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    Name = "${var.env_id}-concourse-lb-security-group"
  }
}

resource "aws_security_group" "concourse_lb_internal_security_group" {
  description = "Concourse Internal"
  vpc_id      = "${aws_vpc.vpc.id}"

  ingress {
    security_groups = ["${aws_security_group.concourse_lb_security_group.id}"]
    protocol    = "tcp"
    from_port   = 8080
    to_port     = 8080
  }

  ingress {
    security_groups = ["${aws_security_group.concourse_lb_security_group.id}"]
    protocol    = "tcp"
    from_port   = 2222
    to_port     = 2222
  }

  egress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    Name = "${var.env_id}-concourse-lb-internal-security-group"
  }
}

output "concourse_lb_internal_security_group" {
  value="${aws_security_group.concourse_lb_internal_security_group.id}"
}

resource "aws_elb" "concourse_lb" {
  name                      = "${var.short_env_id}-concourse-lb"
  cross_zone_load_balancing = true

  health_check {
    healthy_threshold   = 2
    unhealthy_threshold = 10
    interval            = 30
    target              = "TCP:8080"
    timeout             = 5
  }

  listener {
    instance_port     = 8080
    instance_protocol = "tcp"
    lb_port           = 80
    lb_protocol       = "tcp"
  }

  listener {
    instance_port      = 2222
    instance_protocol  = "tcp"
    lb_port            = 2222
    lb_protocol        = "tcp"
  }

  listener {
    instance_port      = 8080
    instance_protocol  = "tcp"
    lb_port            = 443
    lb_protocol        = "ssl"
    ssl_certificate_id = "${aws_iam_server_certificate.lb_cert.arn}"
  }

  security_groups = ["${aws_security_group.concourse_lb_security_group.id}"]
  subnets         = ["${aws_subnet.lb_subnets.*.id}"]
}

output "concourse_lb_name" {
  value = "${aws_elb.concourse_lb.name}"
}

output "concourse_lb_url" {
  value = "${aws_elb.concourse_lb.dns_name}"
}
