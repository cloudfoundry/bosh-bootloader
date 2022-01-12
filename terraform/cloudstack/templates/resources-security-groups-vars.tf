variable "cf_external_ports" {
  type = list(string)
  default = [
    80,
    443,
    2222,
    4443,
  ]
}

variable "bosh_ports" {
  type = list(string)
  default = [
    22,
    6868,
    2555,
    4222,
    25250,
  ]
}

variable "bosh_external_ports" {
  type = list(string)
  default = [
    22,
    6868,
    25555,
    3389,
    8844,
    8443,
  ]
}

variable "shared_tcp_ports" {
  type = list(string)
  default = [
    1801,
    3000,
    3457,
    4003,
    4103,
    4222,
    4443,
    8080,
    8082,
    8300,
    8301,
    8302,
    8443,
    8844,
    8853,
    8889,
    8891,
    9022,
    9023,
    9090,
    9091,
  ]
}

variable "shared_udp_ports" {
  type = list(string)
  default = [
    8301,
    8302,
    8600,
  ]
}
