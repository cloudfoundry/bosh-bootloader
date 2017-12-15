resource "google_service_account" "master" {
  account_id   = "${var.env_id}kubo-master"
  display_name = "${var.env_id} kubo-master"
}

resource "google_service_account" "worker" {
  account_id   = "${var.env_id}kubo-worker"
  display_name = "${var.env_id} kubo-worker"
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
    ]
  }

  binding {
    role = "roles/compute.networkAdmin"

    members = [
      "serviceAccount:${google_service_account.master.email}",
    ]
  }

  binding {
    role = "roles/compute.securityAdmin"

    members = [
      "serviceAccount:${google_service_account.master.email}",

    ]
  }

  binding {
    role = "roles/compute.instanceAdmin"

    members = [
      "serviceAccount:${google_service_account.master.email}",
    ]
  }

  binding {
    role = "roles/iam.serviceAccountActor"

    members = [
      "serviceAccount:${google_service_account.master.email}",
    ]
  }

  binding {
    role = "roles/compute.viewer"

    members = [
      "serviceAccount:${google_service_account.worker.email}",
    ]
  }
}
// Static IP address for HTTP forwarding rule
resource "google_compute_address" "kubo-tcp" {
  name = "${var.env_id}kubo"
}

// TCP Load Balancer
resource "google_compute_target_pool" "kubo-tcp-public" {
    region = "${var.region}"
    name = "${var.env_id}kubo-tcp-public"
}

resource "google_compute_forwarding_rule" "kubo-tcp" {
  name        = "${var.env_id}kubo-tcp"
  target      = "${google_compute_target_pool.kubo-tcp-public.self_link}"
  port_range  = "8443"
  ip_protocol = "TCP"
  ip_address  = "${google_compute_address.kubo-tcp.address}"
}

resource "google_compute_firewall" "kubo-tcp-public" {
  name    = "${var.env_id}kubo-tcp-public"
  network       = "${var.env_id}-network"

  allow {
    protocol = "tcp"
    ports    = ["8443"]
  }

  target_tags = ["master"]
}

output "kubo_master_target_pool" {
   value = "${google_compute_target_pool.kubo-tcp-public.name}"
}

output "master_lb_ip_address" {
  value = "${google_compute_address.kubo-tcp.address}"
}
