variable "elb_idle_timeout" {
  type    = number
  default = 60
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

output "cf_tcp_lb_security_group" {
  value = aws_security_group.cf_tcp_lb_security_group.id
}

output "cf_tcp_lb_internal_security_group" {
  value = aws_security_group.cf_tcp_lb_internal_security_group.id
}

output "cf_router_lb_internal_security_group" {
  value = aws_security_group.cf_router_lb_internal_security_group.id
}

output "cf_router_lb_security_group" {
  value = aws_security_group.cf_router_lb_security_group.id
}

output "cf_ssh_lb_internal_security_group" {
  value = aws_security_group.cf_ssh_lb_internal_security_group.id
}


output "cf_ssh_lb_security_group" {
  value = aws_security_group.cf_ssh_lb_security_group.id
}

