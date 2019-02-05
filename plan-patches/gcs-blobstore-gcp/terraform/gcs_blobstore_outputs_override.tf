output "gcs_blobstore_buildpacks_bucket" {
  value = "${google_storage_bucket.buildpacks.name}"
}

output "gcs_blobstore_droplets_bucket" {
  value = "${google_storage_bucket.droplets.name}"
}

output "gcs_blobstore_packages_bucket" {
  value = "${google_storage_bucket.packages.name}"
}

output "gcs_blobstore_resources_bucket" {
  value = "${google_storage_bucket.resources.name}"
}

output "gcs_blobstore_service_account_project" {
  value = "${google_service_account.blobstore_service_account.project}"
}

output "gcs_blobstore_service_account_email" {
  value = "${google_service_account.blobstore_service_account.email}"
}

output "gcs_blobstore_service_account_key" {
  value     = "${base64decode(google_service_account_key.blobstore_service_account_key.private_key)}"
  sensitive = true
}
