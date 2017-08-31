resource "aws_subnet" "iso1_subnets" {
  count             = "${length(var.availability_zones)}"
  vpc_id            = "${aws_vpc.vpc.id}"
  cidr_block        = "${cidrsubnet("10.0.200.0/24", 4, count.index)}"
  availability_zone = "${element(var.availability_zones, count.index)}"

  tags {
    Name = "${var.env_id}-iso1-subnet${count.index}"
  }
}

output "iso1_az_subnet_id_mapping" {
  value = "${
	  zipmap("${aws_subnet.iso1_subnets.*.availability_zone}", "${aws_subnet.iso1_subnets.*.id}")
	}"
}

resource "aws_elb" "iso1_router_lb" {
  name                      = "bbl-iso1-router-lb" #make this more unique later
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
  subnets         = ["${aws_subnet.iso1_subnets.*.id}"]
}

output "cf_iso1_router_lb_name" {
  value="${aws_elb.iso1_router_lb.name}"
}

resource "aws_route53_record" "iso1_dns" {
  zone_id = "${aws_route53_zone.env_dns_zone.id}"
  name    = "*.iso-seg.${var.system_domain}"
  type    = "CNAME"
  ttl     = 300

  records = ["${aws_elb.iso1_router_lb.dns_name}"]
}

resource "aws_security_group" "iso1_security_group" {
  description = "iso1"
  vpc_id      = "${aws_vpc.vpc.id}"

  ingress {
    self = true
    from_port = 0
    to_port = 0
    protocol = "-1"
  }

  tags {
    Name = "${var.env_id}-iso1-security-group"
  }
}

output "iso1_security_group_id" {
  value="${aws_security_group.iso1_security_group.id}"
}

#iso_shared_security group needs to be attached to primary subnet in the cloud config
resource "aws_security_group" "iso_shared_security_group" {
  description = "iso-shared"
  vpc_id      = "${aws_vpc.vpc.id}"

  ingress {
    self = true
    from_port = 0
    to_port = 0
    protocol = "-1"
  }

  tags {
    Name = "${var.env_id}-iso-shared-security-group"
  }
}

output "iso_shared_security_group_id" {
  value="${aws_security_group.iso_shared_security_group.id}"
}

variable "iso_to_bosh_ports" {
  type = "list"
  default = [22,6868,25555,4222,25250]
}

resource "aws_security_group_rule" "isolation_segments_to_bosh_rule" {
  count = "${length(var.iso_to_bosh_ports)}"
  security_group_id = "${aws_security_group.bosh_security_group.id}"
  type = "ingress"
  protocol = "tcp"
  to_port = "${element(var.iso_to_bosh_ports, count.index)}"
  from_port = "${element(var.iso_to_bosh_ports, count.index)}"
  source_security_group_id = "${aws_security_group.iso1_security_group.id}"
}

variable "iso_to_shared_tcp_ports" {
  type = "list"
  default = [9090,9091,8082,8300,8301,8889,8443,3000,4443,8080,3457,9023,9022,4222]
}

resource "aws_security_group_rule" "isolation_segments_to_shared_tcp_rule" {
  count = "${length(var.iso_to_shared_tcp_ports)}"
  security_group_id = "${aws_security_group.iso_shared_security_group.id}"
  type = "ingress"
  protocol = "tcp"
  to_port = "${element(var.iso_to_shared_tcp_ports, count.index)}"
  from_port = "${element(var.iso_to_shared_tcp_ports, count.index)}"
  source_security_group_id = "${aws_security_group.iso1_security_group.id}"
}

variable "iso_to_shared_udp_ports" {
  type = "list"
  default = [8301,8302,8600]
}

resource "aws_security_group_rule" "isolation_segments_to_shared_udp_rule" {
  count = "${length(var.iso_to_shared_udp_ports)}"
  security_group_id = "${aws_security_group.iso_shared_security_group.id}"
  type = "ingress"
  protocol = "udp"
  to_port = "${element(var.iso_to_shared_udp_ports, count.index)}"
  from_port = "${element(var.iso_to_shared_udp_ports, count.index)}"
  source_security_group_id = "${aws_security_group.iso1_security_group.id}"
}

resource "aws_security_group_rule" "shared_diego_bbs_to_isolated_cells_rule" {
  security_group_id = "${aws_security_group.iso1_security_group.id}"
  type = "ingress"
  protocol = "tcp"
  to_port = 1801
  from_port = 1801
  source_security_group_id = "${aws_security_group.iso_shared_security_group.id}"
}
