variable "kubernetes_master_host" {
  type = "string"
}

resource "google_dns_managed_zone" "cfcr_dns_zone" {
  name        = "${var.env_id}-cfcr-zone"
  dns_name    = "${var.kubernetes_master_host}."
  description = "DNS zone for the ${var.env_id} cfcr environment"
}

resource "google_dns_record_set" "cfcr_api_dns" {
  name       = "${google_dns_managed_zone.cfcr_dns_zone.dns_name}"
  depends_on = ["google_compute_address.cfcr_tcp"]
  type       = "A"
  ttl        = 300

  managed_zone = "${google_dns_managed_zone.cfcr_dns_zone.name}"

  rrdatas = ["${google_compute_address.cfcr_tcp.address}"]
}

