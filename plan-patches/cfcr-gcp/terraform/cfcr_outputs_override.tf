output "cfcr_master_target_pool" {
  value = "${google_compute_target_pool.cfcr_tcp_public.name}"
}

output "cfcr_master_service_account_address" {
  value = "${google_service_account.master.email}"
}

output "cfcr_worker_service_account_address" {
  value = "${google_service_account.worker.email}"
}

output "api-hostname" {
  value = "${google_compute_address.cfcr_tcp.address}"
}

output "project_id" {
  value = "${var.project_id}"
}
