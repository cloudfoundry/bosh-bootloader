resource "google_compute_address" "jumpbox-ip" {
  name = "${var.env_id}-jumpbox-ip"
}

output "jumpbox_url" {
  value = "${google_compute_address.jumpbox-ip.address}:22"
}

output "external_ip" {
  value = google_compute_address.jumpbox-ip.address
}

output "director_address" {
  value = "https://${google_compute_address.jumpbox-ip.address}:25555"
}
