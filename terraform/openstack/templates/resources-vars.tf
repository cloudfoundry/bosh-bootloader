variable "env_id" {}

variable "ext_net_id" {
  description = "OpenStack external network id to create router interface port"
}

variable "availability_zone" {
  description = "OpenStack availability zone name"
}

variable "dns_nameservers" {
  description = "List of DNS server IPs"
  default     = ["8.8.8.8"]
  type        = list(string)
}

variable "ext_net_name" {
  description = "OpenStack external network name to register floating IP"
}

variable "region_name" {
  description = "OpenStack region name"
}
