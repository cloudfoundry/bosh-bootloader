resource "random_id" "kubernetes_cluster_tag" {
  byte_length = 16
}

resource "aws_subnet" "bosh_subnet" {
  tags {
    Name              = "${var.env_id}-bosh-subnet"
    KubernetesCluster = "${random_id.kubernetes_cluster_tag.b64}"
  }
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
  subnets         = ["${aws_subnet.bosh_subnet.id}"]
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
