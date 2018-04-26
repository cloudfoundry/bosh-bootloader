output "kubernetes-cluster-tag" {
  value = "${random_id.kubernetes_cluster_tag.b64}"
}

output "cfcr_master_target_pool" {
   value = "${aws_elb.cfcr_api.name}"
}

output "kubernetes_master_host" {
  value = "${var.kubernetes_master_host}"
}

output "master_iam_instance_profile" {
  value = "${aws_iam_instance_profile.cfcr_master.name}"
}

output "worker_iam_instance_profile" {
  value = "${aws_iam_instance_profile.cfcr_worker.name}"
}
