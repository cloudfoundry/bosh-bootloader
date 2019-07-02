resource "google_service_account" "master" {
  account_id   = "${substr(var.env_id, 0, min(length(var.env_id), 10))}-cfcr-master"
  display_name = "${var.env_id} cfcr-master"
}

resource "google_service_account" "worker" {
  account_id   = "${substr(var.env_id, 0, min(length(var.env_id), 10))}-cfcr-worker"
  display_name = "${var.env_id} cfcr-worker"
}

resource "google_project_iam_member" "storageAdminMaster" {
  project     = "${var.project_id}"
  role = "roles/compute.storageAdmin"
  member = "serviceAccount:${google_service_account.master.email}"
}
resource "google_project_iam_member" "storageAdminWorker" {
  project     = "${var.project_id}"
  role = "roles/compute.storageAdmin"
  member = "serviceAccount:${google_service_account.worker.email}"
}

resource "google_project_iam_member" "objectViewerMaster" {
  project     = "${var.project_id}"
  role = "roles/storage.objectViewer"
  member = "serviceAccount:${google_service_account.master.email}"
}
resource "google_project_iam_member" "objectViewerWorker" {
  project     = "${var.project_id}"
  role = "roles/storage.objectViewer"
  member = "serviceAccount:${google_service_account.worker.email}"
}

resource "google_project_iam_member" "networkAdminMaster" {
  project     = "${var.project_id}"
  role = "roles/compute.networkAdmin"
  member = "serviceAccount:${google_service_account.master.email}"
}
resource "google_project_iam_member" "networkAdminWorker" {
  project     = "${var.project_id}"
  role = "roles/compute.networkAdmin"
  member = "serviceAccount:${google_service_account.worker.email}"
}

resource "google_project_iam_member" "securityAdminMaster" {
  project     = "${var.project_id}"
  role = "roles/compute.securityAdmin"
  member = "serviceAccount:${google_service_account.master.email}"
}
resource "google_project_iam_member" "securityAdminWorker" {
  project     = "${var.project_id}"
  role = "roles/compute.securityAdmin"
  member = "serviceAccount:${google_service_account.worker.email}"
}

resource "google_project_iam_member" "instanceAdminMaster" {
  project     = "${var.project_id}"
  role = "roles/compute.instanceAdmin.v1"
  member = "serviceAccount:${google_service_account.master.email}"
}
resource "google_project_iam_member" "instanceAdminWorker" {
  project     = "${var.project_id}"
  role = "roles/compute.instanceAdmin.v1"
  member = "serviceAccount:${google_service_account.worker.email}"
}

resource "google_project_iam_member" "serviceAccountActorMaster" {
  project     = "${var.project_id}"
  role = "roles/iam.serviceAccountActor"
  member = "serviceAccount:${google_service_account.master.email}"
}
resource "google_project_iam_member" "serviceAccountActorWorker" {
  project     = "${var.project_id}"
  role = "roles/iam.serviceAccountActor"
  member = "serviceAccount:${google_service_account.worker.email}"
}
