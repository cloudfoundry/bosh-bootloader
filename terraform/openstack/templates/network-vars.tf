# input variables

variable "ext_net_id" {
  description = "OpenStack external network id to create router interface port"
}

variable "availability_zone" {
  description = "OpenStack availability zone name"
}

variable "dns_nameservers" {
  description = "List of DNS server IPs"
  default = ["8.8.8.8"]
  type = "list"
}
