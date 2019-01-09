data "template_file" "blobstore_access" {
  template = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:*"
      ],
      "Resource": [
        "$${buildpacks_bucket_arn}",
        "$${buildpacks_bucket_arn}/*",
        "$${droplets_bucket_arn}",
        "$${droplets_bucket_arn}/*",
        "$${packages_bucket_arn}",
        "$${packages_bucket_arn}/*",
        "$${resources_bucket_arn}",
        "$${resources_bucket_arn}/*"
      ]
    }
  ]
}
EOF

  vars {
    buildpacks_bucket_arn = "${aws_s3_bucket.buildpacks_bucket.arn}"
    droplets_bucket_arn   = "${aws_s3_bucket.droplets_bucket.arn}"
    packages_bucket_arn   = "${aws_s3_bucket.packages_bucket.arn}"
    resources_bucket_arn  = "${aws_s3_bucket.resources_bucket.arn}"
  }
}

resource "aws_iam_user" "blobstore_access" {
  name = "${var.env_id}-s3-blobstore-access"
}

resource "aws_iam_access_key" "blobstore_access" {
  user = "${aws_iam_user.blobstore_access.name}"
}

resource "aws_iam_user_policy" "blobstore_access" {
  name = "${var.env_id}-s3-blobstore-access"
  user = "${aws_iam_user.blobstore_access.name}"

  policy = "${data.template_file.blobstore_access.rendered}"
}
