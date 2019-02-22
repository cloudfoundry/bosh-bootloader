resource "random_pet" "environment" {
  count = "${var.cf_env_count}"
}

resource "google_dns_record_set" "wildcard-dns" {
  name       = "*.cf${count.index}.${google_dns_managed_zone.env_dns_zone.dns_name}"
  depends_on = ["google_compute_global_address.cf-address"]
  type       = "A"
  ttl        = 300

  managed_zone = "${google_dns_managed_zone.env_dns_zone.name}"

  rrdatas = ["${element(google_compute_global_address.cf-address.*.address, count.index)}"]

  count = "${var.cf_env_count}"
}

resource "google_dns_record_set" "bosh-dns" {
  name       = "bosh.cf${count.index}.${google_dns_managed_zone.env_dns_zone.dns_name}"
  depends_on = ["google_compute_address.jumpbox-ip"]
  type       = "A"
  ttl        = 300

  managed_zone = "${google_dns_managed_zone.env_dns_zone.name}"

  rrdatas = ["${element(google_compute_address.jumpbox-ip.*.address, count.index)}"]

  count = "${var.cf_env_count}"
}

resource "google_dns_record_set" "cf-ssh-proxy" {
  name       = "ssh.cf${count.index}.${google_dns_managed_zone.env_dns_zone.dns_name}"
  depends_on = ["google_compute_address.cf-ssh-proxy"]
  type       = "A"
  ttl        = 300

  managed_zone = "${google_dns_managed_zone.env_dns_zone.name}"

  rrdatas = ["${element(google_compute_address.cf-ssh-proxy.*.address, count.index)}"]

  count = "${var.cf_env_count}"
}

resource "google_dns_record_set" "tcp-dns" {
  name       = "tcp.cf${count.index}.${google_dns_managed_zone.env_dns_zone.dns_name}"
  depends_on = ["google_compute_address.cf-tcp-router"]
  type       = "A"
  ttl        = 300

  managed_zone = "${google_dns_managed_zone.env_dns_zone.name}"

  rrdatas = ["${element(google_compute_address.cf-tcp-router.*.address, count.index)}"]

  count = "${var.cf_env_count}"
}

resource "google_dns_record_set" "doppler-dns" {
  name       = "doppler.cf${count.index}.${google_dns_managed_zone.env_dns_zone.dns_name}"
  depends_on = ["google_compute_address.cf-ws"]
  type       = "A"
  ttl        = 300

  managed_zone = "${google_dns_managed_zone.env_dns_zone.name}"

  rrdatas = ["${element(google_compute_address.cf-ws.*.address, count.index)}"]

  count = "${var.cf_env_count}"
}

resource "google_dns_record_set" "loggregator-dns" {
  name       = "loggregator.cf${count.index}.${google_dns_managed_zone.env_dns_zone.dns_name}"
  depends_on = ["google_compute_address.cf-ws"]
  type       = "A"
  ttl        = 300

  managed_zone = "${google_dns_managed_zone.env_dns_zone.name}"

  rrdatas = ["${element(google_compute_address.cf-ws.*.address, count.index)}"]

  count = "${var.cf_env_count}"
}

resource "google_dns_record_set" "wildcard-ws-dns" {
  name       = "*.ws.cf${count.index}.${google_dns_managed_zone.env_dns_zone.dns_name}"
  depends_on = ["google_compute_address.cf-ws"]
  type       = "A"
  ttl        = 300

  managed_zone = "${google_dns_managed_zone.env_dns_zone.name}"

  rrdatas = ["${element(google_compute_address.cf-ws.*.address, count.index)}"]

  count = "${var.cf_env_count}"
}
