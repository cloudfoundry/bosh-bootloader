variable "cloudstack_zone" {
  type = string
}

variable "network_vpc_offering" {
  type    = string
  default = "DefaultIsolatedNetworkOfferingForVpcNetworks"
}

variable "vpc_offering" {
  type    = string
  default = "Default VPC offering"
}

variable "cloudstack_compute_offering" {
  type    = string
  default = "shared.large"
}

variable "env_id" {
  type = string
}

variable "short_env_id" {
  type = string
}

variable "vpc_cidr" {
  type    = string
  default = "10.0.0.0/16"
}

variable "dns" {
  type = list(string)
  default = [
    "8.8.8.8",
  ]
}

variable "secure" {
  default = false
}

variable "iso_segment" {
  default = false
}
