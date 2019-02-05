resource "google_service_account" "blobstore_service_account" {
  account_id   = "${var.env_id}-blobstore"
  display_name = "${var.env_id} Blobstore Service Account"
}

resource "google_service_account_key" "blobstore_service_account_key" {
  service_account_id = "${google_service_account.blobstore_service_account.id}"
}

resource "google_storage_bucket_iam_member" "buildpacks_blobstore_cloud_storage_admin" {
  bucket = "${google_storage_bucket.buildpacks.name}"
  role   = "roles/storage.admin"
  member = "serviceAccount:${google_service_account.blobstore_service_account.email}"
}

resource "google_storage_bucket_iam_member" "droplets_blobstore_cloud_storage_admin" {
  bucket = "${google_storage_bucket.droplets.name}"
  role   = "roles/storage.admin"
  member = "serviceAccount:${google_service_account.blobstore_service_account.email}"
}

resource "google_storage_bucket_iam_member" "packages_blobstore_cloud_storage_admin" {
  bucket = "${google_storage_bucket.packages.name}"
  role   = "roles/storage.admin"
  member = "serviceAccount:${google_service_account.blobstore_service_account.email}"
}

resource "google_storage_bucket_iam_member" "resources_blobstore_cloud_storage_admin" {
  bucket = "${google_storage_bucket.resources.name}"
  role   = "roles/storage.admin"
  member = "serviceAccount:${google_service_account.blobstore_service_account.email}"
}
