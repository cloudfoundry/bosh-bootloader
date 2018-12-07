output "cfcr_master_target_pool" {
  value = "${google_compute_target_pool.cfcr_tcp_public.name}"
}

output "cfcr_master_service_account_address" {
  value = "${google_service_account.master.email}"
}

output "cfcr_worker_service_account_address" {
  value = "${google_service_account.worker.email}"
}

output "kubernetes_master_host" {
  value = "${var.kubernetes_master_host}"
}

output "project_id" {
  value = "${var.project_id}"
}
