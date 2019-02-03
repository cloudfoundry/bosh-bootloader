
# access coordinates/credentials
variable "auth_url" {
  description = "Authentication endpoint URL for OpenStack provider (only scheme+host+port, but without path!)"
}

variable "user_name" {
  description = "OpenStack pipeline technical user name"
}

variable "password" {
  description = "OpenStack user password"
}

variable "domain_name" {
  description = "OpenStack domain name"
}

variable "tenant_name" {
  description = "OpenStack project/tenant name"
}

variable "insecure" {
  description = "Skip SSL verification"
  default = "false"
}

variable "cacert_file" {
  description = "Custom CA certificate"
  default = ""
}

# external network coordinates
variable "ext_net_name" {
  description = "OpenStack external network name to register floating IP"
}

# region/zone coordinates
variable "region_name" {
  description = "OpenStack region name"
}