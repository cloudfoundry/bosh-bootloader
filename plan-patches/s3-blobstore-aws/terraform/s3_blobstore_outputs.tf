output "buildpack_directory_key" {
  value = "${aws_s3_bucket.buildpacks_bucket.bucket}"
}

output "droplet_directory_key" {
  value = "${aws_s3_bucket.droplets_bucket.bucket}"
}

output "app_package_directory_key" {
  value = "${aws_s3_bucket.packages_bucket.bucket}"
}

output "resource_directory_key" {
  value = "${aws_s3_bucket.resources_bucket.bucket}"
}

output "blobstore_access_key_id" {
  value = "${aws_iam_access_key.blobstore_access.id}"
}

output "blobstore_secret_access_key" {
  value = "${aws_iam_access_key.blobstore_access.secret}"
  sensitive = true
}

output "aws_region" {
  value = "${var.region}"
}