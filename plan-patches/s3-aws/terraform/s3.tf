variable "s3_bucket_names" {
  type = "list"
}

variable "s3_users" {
  type = "list"
}

variable "s3_replication_region" {
  type = "string"
}

provider "aws" {
  alias      = "s3_replication_provider"
  access_key = "${var.access_key}"
  secret_key = "${var.secret_key}"
  region     = "${var.s3_replication_region}"
}

resource "aws_iam_role" "replication" {
  name = "${var.env_id}-s3-replication-role"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "s3.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_user" "bbl_user" {
  count = "${length(var.s3_users)}"
  name  = "${element(var.s3_users, count.index)}"
}

resource "aws_iam_policy" "replication" {
  name       = "${var.env_id}-s3-replication-policy"
  depends_on = ["aws_s3_bucket.zbucket"]

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:GetReplicationConfiguration",
        "s3:ListBucket"
      ],
      "Effect": "Allow",
      "Resource": [
      "${aws_s3_bucket.zbucket.0.arn}", "${aws_s3_bucket.zbucket.1.arn}","${aws_s3_bucket.zbucket.2.arn}","${aws_s3_bucket.zbucket.3.arn}"
      ]
    },
    {
      "Action": [
        "s3:GetObjectVersion",
        "s3:GetObjectVersionAcl"
      ],
      "Effect": "Allow",
      "Resource": [
      "${aws_s3_bucket.zbucket.0.arn}/*", "${aws_s3_bucket.zbucket.1.arn}/*","${aws_s3_bucket.zbucket.2.arn}/*","${aws_s3_bucket.zbucket.3.arn}/*"
      ]
    },
    {
      "Action": [
        "s3:ReplicateObject",
        "s3:ReplicateDelete"
      ],
      "Effect": "Allow",
      "Resource": [
      "${aws_s3_bucket.destination.0.arn}/*", "${aws_s3_bucket.destination.1.arn}/*", "${aws_s3_bucket.destination.2.arn}/*","${aws_s3_bucket.destination.3.arn}/*"
     ]
    }
  ]
}
POLICY
}

resource "aws_s3_bucket_policy" "bucketpol" {
  count      = "${length(var.s3_bucket_names)}"
  bucket     = "${element(var.s3_bucket_names, count.index)}"
  depends_on = ["aws_s3_bucket.zbucket"]

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "MYBUCKETPOLICY_{element(var.s3_bucket_names, count.index)}",
  "Statement": [
    {
      "Sid": "statementi_${element(var.s3_bucket_names, count.index)}",
      "Effect": "Allow",
      "Principal": {
      "AWS": ${jsonencode(aws_iam_user.bbl_user.*.arn)}
       },
      "Action": "s3:*",
      "Resource": "arn:aws:s3:::${element(var.s3_bucket_names, count.index)}"
     }
  ]
}
POLICY
}

resource "aws_iam_policy_attachment" "replication" {
  name       = "${var.env_id}-s3-replication-policy-attachment"
  roles      = ["${aws_iam_role.replication.name}"]
  policy_arn = "${aws_iam_policy.replication.arn}"
}

resource "aws_s3_bucket" "destination" {
  count    = "${length(var.s3_bucket_names)}"
  provider = "aws.s3_replication_provider"
  bucket   = "${element(var.s3_bucket_names, count.index)}-replica"
  region   = "${var.s3_replication_region}"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket" "zbucket" {
  count      = "${length(var.s3_bucket_names)}"
  bucket     = "${element(var.s3_bucket_names, count.index)}"
  acl        = "private"
  depends_on = ["aws_s3_bucket.destination"]

  versioning {
    enabled = true
  }

  replication_configuration {
    role = "${aws_iam_role.replication.arn}"

    rules {
      id     = "foobar"
      status = "Enabled"
      prefix = ""

      destination {
        bucket        = "arn:aws:s3:::${element(var.s3_bucket_names, count.index)}-replica"
        storage_class = "STANDARD"
      }
    }
  }
}
