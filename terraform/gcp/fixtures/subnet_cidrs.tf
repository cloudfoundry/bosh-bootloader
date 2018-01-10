output "subnet_cidr_1" {
  value = "${cidrsubnet(var.subnet_cidr, 8, 16)}"
}

output "subnet_cidr_2" {
  value = "${cidrsubnet(var.subnet_cidr, 8, 32)}"
}

output "subnet_cidr_3" {
  value = "${cidrsubnet(var.subnet_cidr, 8, 48)}"
}
