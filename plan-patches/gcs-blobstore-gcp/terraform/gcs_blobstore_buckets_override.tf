resource "google_storage_bucket" "buildpacks" {
  name          = "${var.env_id}-buildpacks-bucket"
  force_destroy = true
}

resource "google_storage_bucket" "droplets" {
  name          = "${var.env_id}-droplets-bucket"
  force_destroy = true
}

resource "google_storage_bucket" "packages" {
  name          = "${var.env_id}-packages-bucket"
  force_destroy = true
}

resource "google_storage_bucket" "resources" {
  name          = "${var.env_id}-resources-bucket"
  force_destroy = true
}
