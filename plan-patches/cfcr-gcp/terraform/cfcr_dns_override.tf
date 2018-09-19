variable "kubernetes_master_host"{
  type = "string"
}

variable "custom_managed_zone" {
  type = "string"
  default = ""
}

resource "google_dns_managed_zone" "cfcr_dns_zone" {
  name        = "${var.env_id}-cfcr-zone"
  dns_name    = "${var.kubernetes_master_host}."
  description = "DNS zone for the ${var.env_id} cfcr environment"

  count = "${var.custom_managed_zone != "" ? 0 : 1 }"
}

resource "google_dns_record_set" "cfcr_api_dns" {
  name       = "${var.kubernetes_master_host}."
  depends_on = ["google_compute_address.cfcr_tcp"]
  type       = "A"
  ttl        = 300

  managed_zone = "${var.custom_managed_zone != "" ? var.custom_managed_zone : format("%s-cfcr-zone",var.env_id)}"

  rrdatas = ["${google_compute_address.cfcr_tcp.address}"]
}

