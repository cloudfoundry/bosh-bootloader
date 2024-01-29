provider "cloudstack" {
  api_url    = var.cloudstack_endpoint
  api_key    = var.cloudstack_api_key
  secret_key = var.cloudstack_secret_access_key
  timeout    = "900"
}
