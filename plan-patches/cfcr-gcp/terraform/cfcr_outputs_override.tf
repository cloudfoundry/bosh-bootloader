output "cfcr_master_target_pool" {
   value = "${google_compute_target_pool.cfcr-tcp-public.name}"
}

output "master_lb_ip_address" {
  value = "${google_compute_address.cfcr-tcp.address}"
}

output "service_key_master" {
  sensitive = true
  value = "${base64decode(element(concat(google_service_account_key.master.*.private_key, list("")), 0))}"
}

output "service_key_worker" {
  sensitive = true
  value = "${base64decode(element(concat(google_service_account_key.worker.*.private_key, list("")), 0))}"
}

output "gcp_project_id" {
  value = "${var.project_id}"
}
