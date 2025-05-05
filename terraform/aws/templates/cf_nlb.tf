resource "aws_lb" "cf_ssh_lb" {
  name               = "${var.short_env_id}-cf-ssh-lb"
  internal           = false
  load_balancer_type = "network"
  security_groups    = [aws_security_group.cf_ssh_lb_security_group.id]
  subnets            = [for subnet in aws_subnet.lb_subnets : subnet.id]

  enable_deletion_protection       = false
  enable_cross_zone_load_balancing = true

  # idle_timeout = var.elb_idle_timeout
  ip_address_type = "dualstack"

  tags = {
    Name = var.env_id
  }
}

resource "aws_lb" "cf_router_lb" {
  name               = "${var.short_env_id}-cf-router-lb"
  internal           = false
  load_balancer_type = "network"
  security_groups    = [aws_security_group.cf_router_lb_security_group.id]
  subnets            = [for subnet in aws_subnet.lb_subnets : subnet.id]

  enable_deletion_protection       = false
  enable_cross_zone_load_balancing = true

  # idle_timeout = var.elb_idle_timeout
  ip_address_type = "dualstack"

  tags = {
    Name = var.env_id
  }
}

resource "aws_lb" "cf_tcp_lb" {
  name               = "${var.short_env_id}-cf-tcp-lb"
  internal           = false
  load_balancer_type = "network"
  security_groups    = [aws_security_group.cf_tcp_lb_security_group.id]
  subnets            = [for subnet in aws_subnet.lb_subnets : subnet.id]

  enable_deletion_protection       = false
  enable_cross_zone_load_balancing = true

  # idle_timeout = var.elb_idle_timeout
  ip_address_type = "dualstack"

  tags = {
    Name = var.env_id
  }
}

resource "aws_lb_listener" "cf_tcp_lb" {
  for_each = toset([for x in range(1024, 1074, 1) : tostring(x)])

  load_balancer_arn = aws_lb.cf_tcp_lb.arn
  port              = each.value
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.cf_tcp_nlb[each.value].arn
  }

  depends_on = [
    aws_lb_target_group.cf_tcp_nlb
  ]
}

resource "aws_lb_target_group" "cf_tcp_nlb" {
  for_each = toset([for x in range(1024, 1074, 1) : tostring(x)])

  name     = "${var.short_env_id}-cf-tcp-nlb-${each.value}"
  port     = each.value
  protocol = "TCP"
  vpc_id   = local.vpc_id

  health_check {
    healthy_threshold   = 6
    unhealthy_threshold = 3
    interval            = 15
    protocol            = "TCP"
    port                = 80
  }

  tags = {
    Name = "${var.env_id}-${each.value}"
  }
}

resource "aws_lb_target_group" "cf_ssh_nlb" {
  name     = "${var.short_env_id}-cf-ssh-nlb"
  port     = 2222
  protocol = "TCP"
  vpc_id   = local.vpc_id

  health_check {
    healthy_threshold   = 5
    unhealthy_threshold = 2
    interval            = 12
    protocol            = "TCP"
    port                = 2222
  }

  tags = {
    Name = "${var.env_id}"
  }
}


resource "aws_lb_target_group" "cf_router_nlb" {
  name     = "${var.short_env_id}-cf-router-nlb"
  port     = 80
  protocol = "TCP"
  vpc_id   = local.vpc_id

  health_check {
    healthy_threshold   = 5
    unhealthy_threshold = 2
    interval            = 15
    protocol            = "TCP"
    port                = 80
  }

  tags = {
    Name = "${var.env_id}"
  }
}

resource "aws_lb_listener" "cf_ssh" {
  load_balancer_arn = aws_lb.cf_ssh_lb.arn
  port              = "2222"
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.cf_ssh_nlb.arn
  }
}

resource "aws_lb_listener" "cf_router_http" {
  load_balancer_arn = aws_lb.cf_router_lb.arn
  port              = "80"
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.cf_router_nlb.arn
  }
}

resource "aws_lb_listener" "cf_router_https" {
  load_balancer_arn = aws_lb.cf_router_lb.arn
  port              = "443"
  protocol          = "TLS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.lb_cert.arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.cf_router_nlb.arn
  }
}

resource "aws_lb_listener" "cf_router_4443" {
  load_balancer_arn = aws_lb.cf_router_lb.arn
  port              = "4443"
  protocol          = "TLS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.lb_cert.arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.cf_router_nlb.arn
  }
}

output "cf_ssh_lb_name" {
  value = aws_lb.cf_ssh_lb.name
}

output "cf_ssh_lb_url" {
  value = aws_lb.cf_ssh_lb.dns_name
}

output "cf_router_lb_name" {
  value = aws_lb.cf_router_lb.name
}

output "cf_router_lb_url" {
  value = aws_lb.cf_router_lb.dns_name
}

output "cf_tcp_lb_name" {
  value = aws_lb.cf_tcp_lb.name
}

output "cf_tcp_lb_url" {
  value = aws_lb.cf_tcp_lb.dns_name
}
