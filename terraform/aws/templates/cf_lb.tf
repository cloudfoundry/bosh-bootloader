variable "elb_idle_timeout" {
  type    = number
  default = 60
}

resource "aws_security_group" "cf_ssh_lb_security_group" {
  name        = "${var.env_id}-cf-ssh-lb-security-group"
  description = "CF SSH"
  vpc_id      = local.vpc_id

  ingress {
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = var.dualstack ? ["::/0"] : null
    protocol         = "tcp"
    from_port        = 2222
    to_port          = 2222
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = var.dualstack ? ["::/0"] : null
  }

  tags = {
    Name = "${var.env_id}-cf-ssh-lb-security-group"
  }

  lifecycle {
    ignore_changes = [name]
  }
}

output "cf_ssh_lb_security_group" {
  value = aws_security_group.cf_ssh_lb_security_group.id
}

resource "aws_security_group" "cf_ssh_lb_internal_security_group" {
  name        = "${var.env_id}-cf-ssh-lb-internal-security-group"
  description = "CF SSH Internal"
  vpc_id      = local.vpc_id

  ingress {
    security_groups = ["${aws_security_group.cf_ssh_lb_security_group.id}"]
    protocol        = "tcp"
    from_port       = 2222
    to_port         = 2222
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = var.dualstack ? ["::/0"] : null
  }

  tags = {
    Name = "${var.env_id}-cf-ssh-lb-internal-security-group"
  }

  lifecycle {
    ignore_changes = [name]
  }
}

output "cf_ssh_lb_internal_security_group" {
  value = aws_security_group.cf_ssh_lb_internal_security_group.id
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

  idle_timeout = var.elb_idle_timeout

  security_groups = ["${aws_security_group.cf_ssh_lb_security_group.id}"]
  subnets         = flatten(["${aws_subnet.lb_subnets.*.id}"])

  tags = {
    Name = "${var.env_id}"
  }
}

output "cf_ssh_lb_name" {
  value = aws_elb.cf_ssh_lb.name
}

output "cf_ssh_lb_url" {
  value = aws_elb.cf_ssh_lb.dns_name
}

resource "aws_security_group" "cf_router_lb_security_group" {
  name        = "${var.env_id}-cf-router-lb-security-group"
  description = "CF Router"
  vpc_id      = local.vpc_id

  ingress {
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = var.dualstack ? ["::/0"] : null
    protocol         = "tcp"
    from_port        = 80
    to_port          = 80
  }

  ingress {
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = var.dualstack ? ["::/0"] : null
    protocol         = "tcp"
    from_port        = 443
    to_port          = 443
  }

  ingress {
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = var.dualstack ? ["::/0"] : null
    protocol         = "tcp"
    from_port        = 4443
    to_port          = 4443
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = var.dualstack ? ["::/0"] : null
  }

  tags = {
    Name = "${var.env_id}-cf-router-lb-security-group"
  }

  lifecycle {
    ignore_changes = [name]
  }
}

output "cf_router_lb_security_group" {
  value = aws_security_group.cf_router_lb_security_group.id
}

resource "aws_security_group" "cf_router_lb_internal_security_group" {
  name        = "${var.env_id}-cf-router-lb-internal-security-group"
  description = "CF Router Internal"
  vpc_id      = local.vpc_id

  ingress {
    security_groups = ["${aws_security_group.cf_router_lb_security_group.id}"]
    protocol        = "tcp"
    from_port       = 80
    to_port         = 80
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = var.dualstack ? ["::/0"] : null
  }

  tags = {
    Name = "${var.env_id}-cf-router-lb-internal-security-group"
  }

  lifecycle {
    ignore_changes = [name]
  }
}

output "cf_router_lb_internal_security_group" {
  value = aws_security_group.cf_router_lb_internal_security_group.id
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
    ssl_certificate_id = aws_iam_server_certificate.lb_cert.arn
  }

  listener {
    instance_port      = 80
    instance_protocol  = "tcp"
    lb_port            = 4443
    lb_protocol        = "ssl"
    ssl_certificate_id = aws_iam_server_certificate.lb_cert.arn
  }

  idle_timeout = var.elb_idle_timeout

  security_groups = ["${aws_security_group.cf_router_lb_security_group.id}"]
  subnets         = flatten(["${aws_subnet.lb_subnets.*.id}"])

  tags = {
    Name = "${var.env_id}"
  }
}

resource "aws_lb_target_group" "cf_router_4443" {
  name     = "${var.short_env_id}-routertg-4443"
  port     = 4443
  protocol = "TCP"
  vpc_id   = local.vpc_id

  health_check {
    protocol = "TCP"
  }

  tags = {
    Name = "${var.env_id}"
  }
}

output "cf_router_lb_name" {
  value = aws_elb.cf_router_lb.name
}

output "cf_router_lb_url" {
  value = aws_elb.cf_router_lb.dns_name
}

resource "aws_security_group" "cf_tcp_lb_security_group" {
  name        = "${var.env_id}-cf-tcp-lb-security-group"
  description = "CF TCP"
  vpc_id      = local.vpc_id

  ingress {
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = var.dualstack ? ["::/0"] : null
    protocol         = "tcp"
    from_port        = 1024
    to_port          = 1123
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = var.dualstack ? ["::/0"] : null
  }

  tags = {
    Name = "${var.env_id}-cf-tcp-lb-security-group"
  }

  lifecycle {
    ignore_changes = [name]
  }
}

output "cf_tcp_lb_security_group" {
  value = aws_security_group.cf_tcp_lb_security_group.id
}

resource "aws_security_group" "cf_tcp_lb_internal_security_group" {
  name        = "${var.env_id}-cf-tcp-lb-internal-security-group"
  description = "CF TCP Internal"
  vpc_id      = local.vpc_id

  ingress {
    security_groups = ["${aws_security_group.cf_tcp_lb_security_group.id}"]
    protocol        = "tcp"
    from_port       = 1024
    to_port         = 1123
  }

  ingress {
    security_groups = ["${aws_security_group.cf_tcp_lb_security_group.id}"]
    protocol        = "tcp"
    from_port       = 80
    to_port         = 80
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = var.dualstack ? ["::/0"] : null
  }

  tags = {
    Name = "${var.env_id}-cf-tcp-lb-security-group"
  }

  lifecycle {
    ignore_changes = [name]
  }
}

output "cf_tcp_lb_internal_security_group" {
  value = aws_security_group.cf_tcp_lb_internal_security_group.id
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

  dynamic "listener" {
    for_each = range(1024, 1124, 1)

    content {
      instance_port     = listener.value
      instance_protocol = "tcp"
      lb_port           = listener.value
      lb_protocol       = "tcp"
    }
  }

  idle_timeout = var.elb_idle_timeout

  security_groups = ["${aws_security_group.cf_tcp_lb_security_group.id}"]
  subnets         = flatten(["${aws_subnet.lb_subnets.*.id}"])

  tags = {
    Name = "${var.env_id}"
  }
}

output "cf_tcp_lb_name" {
  value = aws_elb.cf_tcp_lb.name
}

output "cf_tcp_lb_url" {
  value = aws_elb.cf_tcp_lb.dns_name
}
