
resource "aws_security_group_rule" "bosh_security_group_rule_http" {
  security_group_id        = "${aws_security_group.bosh_security_group.id}"
  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 80
  to_port                  = 80
  cidr_blocks = ["${var.bosh_inbound_cidr}"]
}

resource "aws_security_group_rule" "bosh_security_group_rule_https" {
  security_group_id        = "${aws_security_group.bosh_security_group.id}"
  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 443
  to_port                  = 443
  cidr_blocks = ["${var.bosh_inbound_cidr}"]
}
