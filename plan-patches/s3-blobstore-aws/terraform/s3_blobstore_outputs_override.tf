output "s3_blobstore_buildpacks_bucket" {
  value = "${aws_s3_bucket.buildpacks_bucket.bucket}"
}

output "s3_blobstore_droplets_bucket" {
  value = "${aws_s3_bucket.droplets_bucket.bucket}"
}

output "s3_blobstore_packages_bucket" {
  value = "${aws_s3_bucket.packages_bucket.bucket}"
}

output "s3_blobstore_resources_bucket" {
  value = "${aws_s3_bucket.resources_bucket.bucket}"
}

output "s3_blobstore_access_key_id" {
  value = "${aws_iam_access_key.blobstore_access.id}"
}

output "s3_blobstore_secret_access_key" {
  value = "${aws_iam_access_key.blobstore_access.secret}"
}
