resource "google_service_account" "master" {
  account_id   = "${var.env_id}-cfcr-master"
  display_name = "${var.env_id} cfcr-master"
}

resource "google_service_account" "worker" {
  account_id   = "${var.env_id}-cfcr-worker"
  display_name = "${var.env_id} cfcr-worker"
}

resource "google_service_account_key" "master" {
  service_account_id   = "${google_service_account.master.id}"
}

resource "google_service_account_key" "worker" {
  service_account_id   = "${google_service_account.worker.id}"
}

resource "google_project_iam_policy" "policy" {
  project     = "${var.project_id}"
  policy_data = "${data.google_iam_policy.admin.policy_data}"
}

data "google_iam_policy" "admin" {
  binding {
    role = "roles/compute.storageAdmin"

    members = [
      "serviceAccount:${google_service_account.master.email}",
      "serviceAccount:${google_service_account.worker.email}",
    ]
  }

  binding {
    role = "roles/compute.networkAdmin"

    members = [
      "serviceAccount:${google_service_account.master.email}",
      "serviceAccount:${google_service_account.worker.email}",
    ]
  }

  binding {
    role = "roles/compute.securityAdmin"

    members = [
      "serviceAccount:${google_service_account.master.email}",
      "serviceAccount:${google_service_account.worker.email}",
    ]
  }

  binding {
    role = "roles/compute.instanceAdmin"

    members = [
      "serviceAccount:${google_service_account.master.email}",
      "serviceAccount:${google_service_account.worker.email}",
    ]
  }

  binding {
    role = "roles/iam.serviceAccountActor"

    members = [
      "serviceAccount:${google_service_account.master.email}",
      "serviceAccount:${google_service_account.worker.email}",
    ]
  }
}
