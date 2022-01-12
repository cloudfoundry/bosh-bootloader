variable "existing_vpc_id" {
  type        = string
  default     = ""
  description = "Optionally use an existing vpc"
}

locals {
  vpc_count = length(var.existing_vpc_id) > 0 ? 0 : 1
  vpc_id    = length(var.existing_vpc_id) > 0 ? var.existing_vpc_id : join(" ", aws_vpc.vpc.*.id)
}

resource "aws_vpc" "vpc" {
  count                = local.vpc_count
  cidr_block           = var.vpc_cidr
  instance_tenancy     = "default"
  enable_dns_hostnames = true

  tags = {
    Name = "${var.env_id}-vpc"
  }
}
