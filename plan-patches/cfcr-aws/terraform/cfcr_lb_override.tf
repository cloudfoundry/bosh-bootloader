resource "random_id" "kubernetes_cluster_tag" {
  byte_length = 16
}

resource "aws_subnet" "bosh_subnet" {
  tags {
    Name              = "${var.env_id}-bosh-subnet"
    KubernetesCluster = "${random_id.kubernetes_cluster_tag.b64}"
  }
}

resource "aws_subnet" "internal_subnets" {
  tags {
    Name              = "${var.env_id}-internal-subnet${count.index}"
    KubernetesCluster = "${random_id.kubernetes_cluster_tag.b64}"
  }
}

resource "aws_subnet" "k8s_lb_subnets" {
  count             = "${length(var.availability_zones)}"
  vpc_id            = "${local.vpc_id}"
  cidr_block        = "${cidrsubnet(var.vpc_cidr, 8, count.index+5)}"
  availability_zone = "${element(var.availability_zones, count.index)}"

  tags {
    Name              = "${var.env_id}-lb-subnet${count.index}"
    KubernetesCluster = "${random_id.kubernetes_cluster_tag.b64}"
  }

  lifecycle {
    ignore_changes = ["cidr_block", "availability_zone"]
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

resource "aws_route_table_association" "route_k8s_lb_subnets" {
  count          = "${length(var.availability_zones)}"
  subnet_id      = "${element(aws_subnet.k8s_lb_subnets.*.id, count.index)}"
  route_table_id = "${aws_route_table.lb_route_table.id}"
}

resource "aws_security_group" "cfcr_api" {
  name   = "${var.short_env_id}-cfcr-api-access"
  vpc_id = "${local.vpc_id}"

  ingress {
    from_port   = "8443"
    to_port     = "8443"
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group_rule" "cfcr_api_to_internal" {
  security_group_id        = "${aws_security_group.internal_security_group.id}"
  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 8443
  to_port                  = 8443
  source_security_group_id = "${aws_security_group.cfcr_api.id}"
}

resource "aws_elb" "cfcr_api" {
  name            = "${var.short_env_id}-cfcr-api"
  subnets         = ["${aws_subnet.k8s_lb_subnets.*.id}"]
  security_groups = ["${aws_security_group.cfcr_api.id}"]

  listener {
    instance_port     = "8443"
    instance_protocol = "tcp"
    lb_port           = "8443"
    lb_protocol       = "tcp"
  }

  health_check {
    healthy_threshold   = 2
    unhealthy_threshold = 2
    timeout             = 2
    target              = "TCP:8443"
    interval            = 5
  }
}
