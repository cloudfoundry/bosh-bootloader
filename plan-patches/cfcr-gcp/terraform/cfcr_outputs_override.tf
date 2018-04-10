output "cfcr_master_target_pool" {
  value = "${google_compute_target_pool.cfcr-tcp-public.name}"
}

output "cfcr_master_service_account_address" {
  value = "${google_service_account.master.email}"
}

output "cfcr_worker_service_account_address" {
  value = "${google_service_account.worker.email}"
}

output "master_lb_ip_address" {
  value = "${google_compute_address.cfcr-tcp.address}"
}

output "gcp_project_id" {
  value = "${var.project_id}"
}
