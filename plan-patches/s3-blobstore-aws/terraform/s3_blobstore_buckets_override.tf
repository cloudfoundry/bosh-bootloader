resource "aws_s3_bucket" "buildpacks_bucket" {
  bucket        = "${var.env_id}-buildpacks-bucket"
  force_destroy = true
}

resource "aws_s3_bucket" "droplets_bucket" {
  bucket        = "${var.env_id}-droplets-bucket"
  force_destroy = true
}

resource "aws_s3_bucket" "packages_bucket" {
  bucket        = "${var.env_id}-packages-bucket"
  force_destroy = true
}

resource "aws_s3_bucket" "resources_bucket" {
  bucket        = "${var.env_id}-resources-bucket"
  force_destroy = true
}
