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
  {{.SSLCertificateNameProperty}}

  certificate_body  = "${var.ssl_certificate}"
  certificate_chain = "${var.ssl_certificate_chain}"
  private_key       = "${var.ssl_certificate_private_key}"

  lifecycle {
    create_before_destroy = true
	{{.IgnoreSSLCertificateProperties}}
  }
}
