variable "env_id" {}

variable "ext_net_id" {
  description = "OpenStack external network id to create router interface port"
}

variable "availability_zone" {
  description = "OpenStack availability zone name"
}

variable "dns_nameservers" {
  description = "List of DNS server IPs"
  default = ["8.8.8.8"]
  type = list
}

variable "ext_net_name" {
  description = "OpenStack external network name to register floating IP"
}

variable "region_name" {
  description = "OpenStack region name"
}

variable "subnet_cidr" {
  description = "CIDR representing IP range for this subnet, IPv4"
  default = "10.0.1.0/24"
}

variable "subnet_allocation_pool_start" {
  description = "The start IP address available for use with DHCP in this subnet, each IP range must be from the same CIDR that the subnet is part of"
  default = "10.0.1.200"
}

variable "subnet_allocation_pool_end" {
  description = "The end IP address available for use with DHCP in this subnet, each IP range must be from the same CIDR that the subnet is part of"
  default = "10.0.1.254"
}

variable "subnet_gateway_ip" {
  description = "Default gateway used by devices in this subnet"
  default = "10.0.1.1"
}
