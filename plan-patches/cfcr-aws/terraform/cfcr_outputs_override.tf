output "kubernetes_cluster_tag" {
  value = "${random_id.kubernetes_cluster_tag.b64}"
}

output "cfcr_master_target_pool" {
   value = "${aws_elb.cfcr_api.name}"
}

output "api-hostname" {
  value = "${aws_elb.cfcr_api.dns_name}"
}

output "master_iam_instance_profile" {
  value = "${aws_iam_instance_profile.cfcr_master.name}"
}

output "worker_iam_instance_profile" {
  value = "${aws_iam_instance_profile.cfcr_worker.name}"
}
