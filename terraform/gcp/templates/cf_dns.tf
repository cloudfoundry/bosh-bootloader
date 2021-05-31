variable "system_domain" {
  type = string
}

resource "google_dns_managed_zone" "env_dns_zone" {
  name        = "${var.env_id}-zone"
  dns_name    = "${var.system_domain}."
  description = "DNS zone for the ${var.env_id} environment"
}

output "system_domain_dns_servers" {
  value = "${google_dns_managed_zone.env_dns_zone.name_servers}"
}

resource "google_dns_record_set" "wildcard-dns" {
  name       = "*.${google_dns_managed_zone.env_dns_zone.dns_name}"
  depends_on = ["google_compute_global_address.cf-address"]
  type       = "A"
  ttl        = 300

  managed_zone = "${google_dns_managed_zone.env_dns_zone.name}"

  rrdatas = ["${google_compute_global_address.cf-address.address}"]
}

resource "google_dns_record_set" "bosh-dns" {
  name       = "bosh.${google_dns_managed_zone.env_dns_zone.dns_name}"
  depends_on = ["google_compute_address.jumpbox-ip"]
  type       = "A"
  ttl        = 300

  managed_zone = "${google_dns_managed_zone.env_dns_zone.name}"

  rrdatas = ["${google_compute_address.jumpbox-ip.address}"]
}

resource "google_dns_record_set" "cf-ssh-proxy" {
  name       = "ssh.${google_dns_managed_zone.env_dns_zone.dns_name}"
  depends_on = ["google_compute_address.cf-ssh-proxy"]
  type       = "A"
  ttl        = 300

  managed_zone = "${google_dns_managed_zone.env_dns_zone.name}"

  rrdatas = ["${google_compute_address.cf-ssh-proxy.address}"]
}

resource "google_dns_record_set" "tcp-dns" {
  name       = "tcp.${google_dns_managed_zone.env_dns_zone.dns_name}"
  depends_on = ["google_compute_address.cf-tcp-router"]
  type       = "A"
  ttl        = 300

  managed_zone = "${google_dns_managed_zone.env_dns_zone.name}"

  rrdatas = ["${google_compute_address.cf-tcp-router.address}"]
}

resource "google_dns_record_set" "doppler-dns" {
  name       = "doppler.${google_dns_managed_zone.env_dns_zone.dns_name}"
  depends_on = ["google_compute_address.cf-ws"]
  type       = "A"
  ttl        = 300

  managed_zone = "${google_dns_managed_zone.env_dns_zone.name}"

  rrdatas = ["${google_compute_address.cf-ws.address}"]
}

resource "google_dns_record_set" "loggregator-dns" {
  name       = "loggregator.${google_dns_managed_zone.env_dns_zone.dns_name}"
  depends_on = ["google_compute_address.cf-ws"]
  type       = "A"
  ttl        = 300

  managed_zone = "${google_dns_managed_zone.env_dns_zone.name}"

  rrdatas = ["${google_compute_address.cf-ws.address}"]
}

resource "google_dns_record_set" "wildcard-ws-dns" {
  name       = "*.ws.${google_dns_managed_zone.env_dns_zone.dns_name}"
  depends_on = ["google_compute_address.cf-ws"]
  type       = "A"
  ttl        = 300

  managed_zone = "${google_dns_managed_zone.env_dns_zone.name}"

  rrdatas = ["${google_compute_address.cf-ws.address}"]
}
