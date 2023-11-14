terraform {
  required_providers {
    cloudstack = {
      source = "orange-cloudfoundry/cloudstack"
    }
    local      = {
      source = "hashicorp/local"
    }
    template   = {
      source = "hashicorp/template"
    }
    tls        = {
      source = "hashicorp/tls"
    }
  }
  required_version = ">= 0.13"
}

provider "cloudstack" {
  api_url    = var.cloudstack_endpoint
  api_key    = var.cloudstack_api_key
  secret_key = var.cloudstack_secret_access_key
  timeout    = "900"
}
