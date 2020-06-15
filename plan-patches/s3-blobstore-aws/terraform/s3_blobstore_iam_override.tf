resource "aws_iam_user" "blobstore_access" {
  name = "${var.env_id}-s3-blobstore-access"
}

resource "aws_iam_access_key" "blobstore_access" {
  user = aws_iam_user.blobstore_access.name
}

resource "aws_iam_user_policy" "blobstore_access" {
  name = "${var.env_id}-s3-blobstore-access"
  user = aws_iam_user.blobstore_access.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:*"
      ],
      "Resource": [
        "${aws_s3_bucket.buildpacks_bucket.arn}",
        "${aws_s3_bucket.buildpacks_bucket.arn}/*",
        "${aws_s3_bucket.droplets_bucket.arn}",
        "${aws_s3_bucket.droplets_bucket.arn}/*",
        "${aws_s3_bucket.packages_bucket.arn}",
        "${aws_s3_bucket.packages_bucket.arn}/*",
        "${aws_s3_bucket.resources_bucket.arn}",
        "${aws_s3_bucket.resources_bucket.arn}/*"
      ]
    }
  ]
}
EOF
}
