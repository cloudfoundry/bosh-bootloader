resource "aws_eip" "bosh_eip" {
  depends_on = ["aws_internet_gateway.ig"]
  vpc      = true
}

output "external_ip" {
  value = "${aws_eip.bosh_eip.public_ip}"
}

output "director_address" {
  value = "https://${aws_eip.bosh_eip.public_ip}:25555"
}

resource "aws_iam_role" "bosh" {
  name = "${var.env_id}_bosh_role"
  path = "/"
  lifecycle {
    create_before_destroy = true
  }

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "bosh" {
  name   = "${var.env_id}_bosh_policy"
  path   = "/"
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
	    "iam:PassRole"
	  ],
	  "Effect": "Allow",
	  "Resource": "${aws_iam_role.bosh.arn}"
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

resource "aws_iam_role_policy_attachment" "bosh" {
  role = "${var.env_id}_bosh_role"
  policy_arn = "${aws_iam_policy.bosh.arn}"
}

resource "aws_iam_instance_profile" "bosh" {
  role = "${aws_iam_role.bosh.name}"
}

output "bosh_iam_instance_profile" {
  value = "${aws_iam_instance_profile.bosh.name}"
}

variable "nat_ami_map" {
  type = "map"

  default = {
    ap-northeast-1 = "ami-f885ae96"
    ap-northeast-2 = "ami-4118d72f"
    ap-southeast-1 = "ami-e2fc3f81"
    ap-southeast-2 = "ami-e3217a80"
    eu-central-1 = "ami-0b322e67"
    eu-west-1 = "ami-c0993ab3"
    sa-east-1 = "ami-8631b5ea"
    us-east-1 = "ami-68115b02"
    us-east-2 = "ami-6893b20d"
    us-west-1 = "ami-ef1a718f"
    us-west-2 = "ami-77a4b816"
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

resource "aws_default_security_group" "default_security_group" {
	vpc_id = "${aws_vpc.vpc.id}"
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

  tags {
    Name = "${var.env_id}-bosh-subnet"
  }
}

resource "aws_route_table" "bosh_route_table" {
  vpc_id = "${aws_vpc.vpc.id}"
}

resource "aws_route" "bosh_route_table" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id = "${aws_internet_gateway.ig.id}"
  route_table_id = "${aws_route_table.bosh_route_table.id}"
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

  lifecycle {
    ignore_changes = ["cidr_block", "availability_zone"]
  }
}

resource "aws_route_table" "internal_route_table" {
  vpc_id = "${aws_vpc.vpc.id}"
}

resource "aws_route" "internal_route_table" {
  destination_cidr_block = "0.0.0.0/0"
  instance_id = "${aws_instance.nat.id}"
  route_table_id = "${aws_route_table.internal_route_table.id}"
}

resource "aws_route_table_association" "route_internal_subnets" {
  count          = "${length(var.availability_zones)}"
  subnet_id      = "${element(aws_subnet.internal_subnets.*.id, count.index)}"
  route_table_id = "${aws_route_table.internal_route_table.id}"
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

  lifecycle {
    ignore_changes = ["cidr_block", "availability_zone"]
  }
}

resource "aws_route_table" "lb_route_table" {
  vpc_id = "${aws_vpc.vpc.id}"
}

resource "aws_route" "lb_route_table" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id = "${aws_internet_gateway.ig.id}"
  route_table_id = "${aws_route_table.lb_route_table.id}"
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

resource "aws_security_group" "cf_ssh_lb_security_group" {
  description = "CF SSH"
  vpc_id      = "${aws_vpc.vpc.id}"

  ingress {
    cidr_blocks = ["0.0.0.0/0"]
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
    Name = "${var.env_id}-cf-ssh-lb-security-group"
  }
}

output "cf_ssh_lb_security_group" {
  value="${aws_security_group.cf_ssh_lb_security_group.id}"
}

resource "aws_security_group" "cf_ssh_lb_internal_security_group" {
  description = "CF SSH Internal"
  vpc_id      = "${aws_vpc.vpc.id}"

  ingress {
    security_groups = ["${aws_security_group.cf_ssh_lb_security_group.id}"]
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
    Name = "${var.env_id}-cf-ssh-lb-internal-security-group"
  }
}

output "cf_ssh_lb_internal_security_group" {
  value="${aws_security_group.cf_ssh_lb_internal_security_group.id}"
}

resource "aws_elb" "cf_ssh_lb" {
  name                      = "${var.short_env_id}-cf-ssh-lb"
  cross_zone_load_balancing = true

  health_check {
    healthy_threshold   = 5
    unhealthy_threshold = 2
    interval            = 6
    target              = "TCP:2222"
    timeout             = 2
  }

  listener {
    instance_port     = 2222
    instance_protocol = "tcp"
    lb_port           = 2222
    lb_protocol       = "tcp"
  }

  security_groups = ["${aws_security_group.cf_ssh_lb_security_group.id}"]
  subnets         = ["${aws_subnet.lb_subnets.*.id}"]
}

output "cf_ssh_lb_name" {
  value = "${aws_elb.cf_ssh_lb.name}"
}

output "cf_ssh_lb_url" {
  value = "${aws_elb.cf_ssh_lb.dns_name}"
}

resource "aws_security_group" "cf_router_lb_security_group" {
  description = "CF Router"
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
    from_port   = 443
    to_port     = 443
  }

  ingress {
    cidr_blocks = ["0.0.0.0/0"]
    protocol    = "tcp"
    from_port   = 4443
    to_port     = 4443
  }

  egress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    Name = "${var.env_id}-cf-router-lb-security-group"
  }
}

output "cf_router_lb_security_group" {
  value="${aws_security_group.cf_router_lb_security_group.id}"
}

resource "aws_security_group" "cf_router_lb_internal_security_group" {
  description = "CF Router Internal"
  vpc_id      = "${aws_vpc.vpc.id}"

  ingress {
    security_groups = ["${aws_security_group.cf_router_lb_security_group.id}"]
    protocol    = "tcp"
    from_port   = 80
    to_port     = 80
  }

  egress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    Name = "${var.env_id}-cf-router-lb-internal-security-group"
  }
}

output "cf_router_lb_internal_security_group" {
  value="${aws_security_group.cf_router_lb_internal_security_group.id}"
}

resource "aws_elb" "cf_router_lb" {
  name                      = "${var.short_env_id}-cf-router-lb"
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
    ssl_certificate_id = "${aws_iam_server_certificate.lb_cert.arn}"
  }

  listener {
    instance_port      = 80
    instance_protocol  = "tcp"
    lb_port            = 4443
    lb_protocol        = "ssl"
    ssl_certificate_id = "${aws_iam_server_certificate.lb_cert.arn}"
  }

  security_groups = ["${aws_security_group.cf_router_lb_security_group.id}"]
  subnets         = ["${aws_subnet.lb_subnets.*.id}"]
}

output "cf_router_lb_name" {
  value = "${aws_elb.cf_router_lb.name}"
}

output "cf_router_lb_url" {
  value = "${aws_elb.cf_router_lb.dns_name}"
}

resource "aws_security_group" "cf_tcp_lb_security_group" {
  description = "CF TCP"
  vpc_id      = "${aws_vpc.vpc.id}"

  ingress {
    cidr_blocks = ["0.0.0.0/0"]
    protocol    = "tcp"
    from_port   = 1024
    to_port     = 1123
  }

  egress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    Name = "${var.env_id}-cf-tcp-lb-security-group"
  }
}

output "cf_tcp_lb_security_group" {
  value="${aws_security_group.cf_tcp_lb_security_group.id}"
}

resource "aws_security_group" "cf_tcp_lb_internal_security_group" {
  description = "CF TCP Internal"
  vpc_id      = "${aws_vpc.vpc.id}"

  ingress {
    security_groups = ["${aws_security_group.cf_tcp_lb_security_group.id}"]
    protocol    = "tcp"
    from_port   = 1024
    to_port     = 1123
  }

  ingress {
    security_groups = ["${aws_security_group.cf_tcp_lb_security_group.id}"]
    protocol    = "tcp"
    from_port   = 80
    to_port     = 80
  }

  egress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    Name = "${var.env_id}-cf-tcp-lb-internal-security-group"
  }
}

output "cf_tcp_lb_internal_security_group" {
  value="${aws_security_group.cf_tcp_lb_internal_security_group.id}"
}

resource "aws_elb" "cf_tcp_lb" {
  name                      = "${var.short_env_id}-cf-tcp-lb"
  cross_zone_load_balancing = true

  health_check {
    healthy_threshold   = 6
    unhealthy_threshold = 3
    interval            = 5
    target              = "TCP:80"
    timeout             = 3
  }

  listener {
    instance_port     = 1024
    instance_protocol = "tcp"
    lb_port           = 1024
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1025
    instance_protocol = "tcp"
    lb_port           = 1025
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1026
    instance_protocol = "tcp"
    lb_port           = 1026
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1027
    instance_protocol = "tcp"
    lb_port           = 1027
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1028
    instance_protocol = "tcp"
    lb_port           = 1028
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1029
    instance_protocol = "tcp"
    lb_port           = 1029
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1030
    instance_protocol = "tcp"
    lb_port           = 1030
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1031
    instance_protocol = "tcp"
    lb_port           = 1031
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1032
    instance_protocol = "tcp"
    lb_port           = 1032
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1033
    instance_protocol = "tcp"
    lb_port           = 1033
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1034
    instance_protocol = "tcp"
    lb_port           = 1034
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1035
    instance_protocol = "tcp"
    lb_port           = 1035
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1036
    instance_protocol = "tcp"
    lb_port           = 1036
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1037
    instance_protocol = "tcp"
    lb_port           = 1037
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1038
    instance_protocol = "tcp"
    lb_port           = 1038
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1039
    instance_protocol = "tcp"
    lb_port           = 1039
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1040
    instance_protocol = "tcp"
    lb_port           = 1040
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1041
    instance_protocol = "tcp"
    lb_port           = 1041
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1042
    instance_protocol = "tcp"
    lb_port           = 1042
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1043
    instance_protocol = "tcp"
    lb_port           = 1043
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1044
    instance_protocol = "tcp"
    lb_port           = 1044
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1045
    instance_protocol = "tcp"
    lb_port           = 1045
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1046
    instance_protocol = "tcp"
    lb_port           = 1046
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1047
    instance_protocol = "tcp"
    lb_port           = 1047
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1048
    instance_protocol = "tcp"
    lb_port           = 1048
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1049
    instance_protocol = "tcp"
    lb_port           = 1049
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1050
    instance_protocol = "tcp"
    lb_port           = 1050
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1051
    instance_protocol = "tcp"
    lb_port           = 1051
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1052
    instance_protocol = "tcp"
    lb_port           = 1052
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1053
    instance_protocol = "tcp"
    lb_port           = 1053
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1054
    instance_protocol = "tcp"
    lb_port           = 1054
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1055
    instance_protocol = "tcp"
    lb_port           = 1055
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1056
    instance_protocol = "tcp"
    lb_port           = 1056
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1057
    instance_protocol = "tcp"
    lb_port           = 1057
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1058
    instance_protocol = "tcp"
    lb_port           = 1058
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1059
    instance_protocol = "tcp"
    lb_port           = 1059
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1060
    instance_protocol = "tcp"
    lb_port           = 1060
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1061
    instance_protocol = "tcp"
    lb_port           = 1061
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1062
    instance_protocol = "tcp"
    lb_port           = 1062
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1063
    instance_protocol = "tcp"
    lb_port           = 1063
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1064
    instance_protocol = "tcp"
    lb_port           = 1064
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1065
    instance_protocol = "tcp"
    lb_port           = 1065
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1066
    instance_protocol = "tcp"
    lb_port           = 1066
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1067
    instance_protocol = "tcp"
    lb_port           = 1067
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1068
    instance_protocol = "tcp"
    lb_port           = 1068
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1069
    instance_protocol = "tcp"
    lb_port           = 1069
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1070
    instance_protocol = "tcp"
    lb_port           = 1070
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1071
    instance_protocol = "tcp"
    lb_port           = 1071
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1072
    instance_protocol = "tcp"
    lb_port           = 1072
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1073
    instance_protocol = "tcp"
    lb_port           = 1073
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1074
    instance_protocol = "tcp"
    lb_port           = 1074
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1075
    instance_protocol = "tcp"
    lb_port           = 1075
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1076
    instance_protocol = "tcp"
    lb_port           = 1076
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1077
    instance_protocol = "tcp"
    lb_port           = 1077
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1078
    instance_protocol = "tcp"
    lb_port           = 1078
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1079
    instance_protocol = "tcp"
    lb_port           = 1079
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1080
    instance_protocol = "tcp"
    lb_port           = 1080
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1081
    instance_protocol = "tcp"
    lb_port           = 1081
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1082
    instance_protocol = "tcp"
    lb_port           = 1082
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1083
    instance_protocol = "tcp"
    lb_port           = 1083
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1084
    instance_protocol = "tcp"
    lb_port           = 1084
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1085
    instance_protocol = "tcp"
    lb_port           = 1085
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1086
    instance_protocol = "tcp"
    lb_port           = 1086
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1087
    instance_protocol = "tcp"
    lb_port           = 1087
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1088
    instance_protocol = "tcp"
    lb_port           = 1088
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1089
    instance_protocol = "tcp"
    lb_port           = 1089
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1090
    instance_protocol = "tcp"
    lb_port           = 1090
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1091
    instance_protocol = "tcp"
    lb_port           = 1091
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1092
    instance_protocol = "tcp"
    lb_port           = 1092
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1093
    instance_protocol = "tcp"
    lb_port           = 1093
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1094
    instance_protocol = "tcp"
    lb_port           = 1094
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1095
    instance_protocol = "tcp"
    lb_port           = 1095
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1096
    instance_protocol = "tcp"
    lb_port           = 1096
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1097
    instance_protocol = "tcp"
    lb_port           = 1097
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1098
    instance_protocol = "tcp"
    lb_port           = 1098
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1099
    instance_protocol = "tcp"
    lb_port           = 1099
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1100
    instance_protocol = "tcp"
    lb_port           = 1100
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1101
    instance_protocol = "tcp"
    lb_port           = 1101
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1102
    instance_protocol = "tcp"
    lb_port           = 1102
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1103
    instance_protocol = "tcp"
    lb_port           = 1103
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1104
    instance_protocol = "tcp"
    lb_port           = 1104
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1105
    instance_protocol = "tcp"
    lb_port           = 1105
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1106
    instance_protocol = "tcp"
    lb_port           = 1106
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1107
    instance_protocol = "tcp"
    lb_port           = 1107
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1108
    instance_protocol = "tcp"
    lb_port           = 1108
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1109
    instance_protocol = "tcp"
    lb_port           = 1109
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1110
    instance_protocol = "tcp"
    lb_port           = 1110
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1111
    instance_protocol = "tcp"
    lb_port           = 1111
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1112
    instance_protocol = "tcp"
    lb_port           = 1112
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1113
    instance_protocol = "tcp"
    lb_port           = 1113
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1114
    instance_protocol = "tcp"
    lb_port           = 1114
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1115
    instance_protocol = "tcp"
    lb_port           = 1115
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1116
    instance_protocol = "tcp"
    lb_port           = 1116
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1117
    instance_protocol = "tcp"
    lb_port           = 1117
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1118
    instance_protocol = "tcp"
    lb_port           = 1118
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1119
    instance_protocol = "tcp"
    lb_port           = 1119
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1120
    instance_protocol = "tcp"
    lb_port           = 1120
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1121
    instance_protocol = "tcp"
    lb_port           = 1121
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1122
    instance_protocol = "tcp"
    lb_port           = 1122
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 1123
    instance_protocol = "tcp"
    lb_port           = 1123
    lb_protocol       = "tcp"
  }

  security_groups = ["${aws_security_group.cf_tcp_lb_security_group.id}"]
  subnets         = ["${aws_subnet.lb_subnets.*.id}"]
}

output "cf_tcp_lb_name" {
  value = "${aws_elb.cf_tcp_lb.name}"
}

output "cf_tcp_lb_url" {
  value = "${aws_elb.cf_tcp_lb.dns_name}"
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

variable "ssl_certificate_name" {
  type = "string"
}

variable "ssl_certificate_name_prefix" {
  type = "string"
}

resource "aws_iam_server_certificate" "lb_cert" {
  name_prefix       = "${var.ssl_certificate_name_prefix}"

  certificate_body  = "${var.ssl_certificate}"
  certificate_chain = "${var.ssl_certificate_chain}"
  private_key       = "${var.ssl_certificate_private_key}"

  lifecycle {
    create_before_destroy = true
	ignore_changes = ["certificate_body", "certificate_chain", "private_key"]
  }
}

variable "system_domain" {
  type = "string"
}

resource "aws_route53_zone" "env_dns_zone" {
  name = "${var.system_domain}"

  tags {
    Name = "${var.env_id}-hosted-zone"
  }
}

output "env_dns_zone_name_servers" {
  value = "${aws_route53_zone.env_dns_zone.name_servers}"
}

resource "aws_route53_record" "wildcard_dns" {
  zone_id = "${aws_route53_zone.env_dns_zone.id}"
  name    = "*.${var.system_domain}"
  type    = "CNAME"
  ttl     = 300

  records = ["${aws_elb.cf_router_lb.dns_name}"]
}

resource "aws_route53_record" "ssh" {
  zone_id = "${aws_route53_zone.env_dns_zone.id}"
  name    = "ssh.${var.system_domain}"
  type    = "CNAME"
  ttl     = 300

  records = ["${aws_elb.cf_ssh_lb.dns_name}"]
}

resource "aws_route53_record" "bosh" {
  zone_id = "${aws_route53_zone.env_dns_zone.id}"
  name    = "bosh.${var.system_domain}"
  type    = "A"
  ttl     = 300

  records = ["${aws_eip.bosh_eip.public_ip}"]
}

resource "aws_route53_record" "tcp" {
  zone_id = "${aws_route53_zone.env_dns_zone.id}"
  name    = "tcp.${var.system_domain}"
  type    = "CNAME"
  ttl     = 300

  records = ["${aws_elb.cf_tcp_lb.dns_name}"]
}
