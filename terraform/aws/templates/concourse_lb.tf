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
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
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
    protocol        = "tcp"
    from_port       = 8080
    to_port         = 8080
  }

  ingress {
    security_groups = ["${aws_security_group.concourse_lb_security_group.id}"]
    protocol        = "tcp"
    from_port       = 2222
    to_port         = 2222
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    Name = "${var.env_id}-concourse-lb-internal-security-group"
  }
}

output "concourse_lb_internal_security_group" {
  value = "${aws_security_group.concourse_lb_internal_security_group.id}"
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
    instance_port     = 2222
    instance_protocol = "tcp"
    lb_port           = 2222
    lb_protocol       = "tcp"
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
