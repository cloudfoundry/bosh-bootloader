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

resource "aws_iam_user" "bbl_s3_user" {
  count = "${length(var.s3_users)}"
  name  = "${element(var.s3_users, count.index)}"
}

resource "aws_kms_key" "kms_key_s3_rep" {
  provider            = "aws.s3_replication_provider"
  enable_key_rotation = true

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Id": "key-consolepolicy-3",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::481316288090:root"
      },
      "Action": "kms:*",
      "Resource": "*"
    },
    {
      "Sid": "Allow access for Key Administrators",
      "Effect": "Allow",
      "Principal": {
        "AWS": ${jsonencode(aws_iam_user.bbl_s3_user.*.arn)}

      },
      "Action": [
        "kms:Create*",
        "kms:Describe*",
        "kms:Enable*",
        "kms:List*",
        "kms:Put*",
        "kms:Update*",
        "kms:Revoke*",
        "kms:Disable*",
        "kms:Get*",
        "kms:Delete*",
        "kms:ScheduleKeyDeletion",
        "kms:CancelKeyDeletion"
      ],
      "Resource": "*"
    },
    {
      "Sid": "Allow use of the key",
      "Effect": "Allow",
      "Principal": {
        "AWS": ${jsonencode(concat(formatlist("%s",aws_iam_user.bbl_s3_user.*.arn),formatlist("%s",aws_iam_role.s3-replication.*.arn)))}
      },
      "Action": [
        "kms:Encrypt",
        "kms:Decrypt",
        "kms:ReEncrypt*",
        "kms:GenerateDataKey*",
        "kms:DescribeKey",
        "kms:Create*",
        "kms:Describe*",
        "kms:Enable*",
        "kms:List*",
        "kms:Put*",
        "kms:Update*",
        "kms:Revoke*",
        "kms:Disable*",
        "kms:Get*",
        "kms:Delete*",
        "kms:ScheduleKeyDeletion",
        "kms:CancelKeyDeletion"
      ],
      "Resource": "*"
    },
    {
      "Sid": "Allow attachment of persistent resources",
      "Effect": "Allow",
      "Principal": {
        "AWS": ${jsonencode(aws_iam_user.bbl_s3_user.*.arn)}
      },
      "Action": [
        "kms:CreateGrant",
        "kms:ListGrants",
        "kms:RevokeGrant"
      ],
      "Resource": "*",
      "Condition": {
        "Bool": {
          "kms:GrantIsForAWSResource": "true"
        }
      }
    }
  ]
}
EOF
}

resource "aws_kms_key" "kms_key_s3" {
  enable_key_rotation = true

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Id": "key-consolepolicy-2",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::481316288090:root"
      },
      "Action": "kms:*",
      "Resource": "*"
    },
    {
      "Sid": "Allow access for Key Administrators",
      "Effect": "Allow",
      "Principal": {
        "AWS": ${jsonencode(concat(formatlist("%s",aws_iam_user.bbl_s3_user.*.arn),formatlist("%s",aws_iam_role.s3-replication.*.arn)))}

      },
      "Action": [
        "kms:Create*",
        "kms:Describe*",
        "kms:Enable*",
        "kms:List*",
        "kms:Put*",
        "kms:Update*",
        "kms:Revoke*",
        "kms:Disable*",
        "kms:Get*",
        "kms:Delete*",
        "kms:ScheduleKeyDeletion",
        "kms:CancelKeyDeletion"
      ],
      "Resource": "*"
    },
    {
      "Sid": "Allow use of the key",
      "Effect": "Allow",
      "Principal": {
        "AWS": ${jsonencode(concat(formatlist("%s",aws_iam_user.bbl_s3_user.*.arn),formatlist("%s",aws_iam_role.s3-replication.*.arn)))}
      },
      "Action": [
        "kms:Encrypt",
        "kms:Decrypt",
        "kms:ReEncrypt*",
        "kms:GenerateDataKey*",
        "kms:DescribeKey"
      ],
      "Resource": "*"
    },
    {
      "Sid": "Allow attachment of persistent resources",
      "Effect": "Allow",
      "Principal": {
        "AWS": ${jsonencode(aws_iam_user.bbl_s3_user.*.arn)}
      },
      "Action": [
        "kms:CreateGrant",
        "kms:ListGrants",
        "kms:RevokeGrant"
      ],
      "Resource": "*",
      "Condition": {
        "Bool": {
          "kms:GrantIsForAWSResource": "true"
        }
      }
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "destinations" {
  count      = "${length(var.s3_bucket_names)}"
  provider   = "aws.s3_replication_provider"
  bucket     = "${element(var.s3_bucket_names, count.index)}-replica"
  region     = "${var.s3_replication_region}"


  versioning {
    enabled = true
  }
}
resource "aws_iam_role" "s3-replication" {
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

resource "aws_iam_policy" "s3-replication" {
  name       = "${var.env_id}-s3-replication-policy"
  depends_on = ["aws_s3_bucket.zbuckets"]

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
      "s3:GetObjectVersionForReplication",
      "s3:GetObjectVersionAcl",
        "s3:GetObjectVersionForReplication",
        "s3:ListBucket",
        "s3:GetReplicationConfiguration",
        "s3:GetObjectVersionForReplication",
        "s3:GetObjectVersionAcl"
      ],
      "Effect": "Allow",
      "Resource":
      ${jsonencode(aws_s3_bucket.zbuckets.*.arn)}
    },
    {
      "Action": [
      "s3:GetObjectVersionForReplication",
      "s3:GetObjectVersionAcl",
        "s3:GetObjectVersion",
        "s3:ListBucket",
        "s3:GetReplicationConfiguration",
        "s3:GetObjectVersionForReplication",
        "s3:GetObjectVersionAcl"
      ],
      "Effect": "Allow",
      "Resource":
      ${jsonencode(concat(formatlist("%s/*",aws_s3_bucket.zbuckets.*.arn)))}
    },
    {
      "Action": [
        "s3:ReplicateObject",
        "s3:ReplicateDelete"
      ],
      "Effect": "Allow",
      "Resource":
      ${jsonencode(concat(formatlist("%s/*",aws_s3_bucket.destinations.*.arn)))}
    },
    {
    "Action": [
            "kms:*"
         ],
    "Effect": "Allow",
    "Resource": "*"
    },
    {
        "Action": [
            "s3:ReplicateObject",
            "s3:ReplicateDelete",
            "s3:ReplicateTags",
            "s3:GetObjectVersionTagging"
        ],
        "Effect": "Allow",
        "Condition": {
            "StringLikeIfExists": {
                "s3:x-amz-server-side-encryption": [
                    "aws:kms",
                    "AES256"
                ],
                "s3:x-amz-server-side-encryption-aws-kms-key-id": [
                    "${aws_kms_key.kms_key_s3_rep.arn}"
                ]
            }
        },
        "Resource": ${jsonencode(concat(formatlist("%s/*",aws_s3_bucket.destinations.*.arn)))}
    },
    {
        "Action": [
            "kms:Decrypt"
        ],
        "Effect": "Allow",
        "Condition": {
            "StringLike": {
                "kms:ViaService": "s3.us-east-1.amazonaws.com",
                "kms:EncryptionContext:aws:s3:arn":
                    ${jsonencode(concat(formatlist("%s/*",aws_s3_bucket.zbuckets.*.arn)))}

            }
        },
        "Resource": [
            "${aws_kms_key.kms_key_s3.arn}"
        ]
    },
    {
        "Action": [
            "kms:Encrypt"
        ],
        "Effect": "Allow",
        "Condition": {
            "StringLike": {
                "kms:ViaService": "s3.us-west-1.amazonaws.com",
                "kms:EncryptionContext:aws:s3:arn":
                    ${jsonencode(concat(formatlist("%s/*",aws_s3_bucket.destinations.*.arn)))}

            }
        },
        "Resource": [
            "${aws_kms_key.kms_key_s3_rep.arn}"
        ]
    }
  ]
}
POLICY
}

resource "aws_s3_bucket_policy" "bucketpol" {
  count      = "${length(var.s3_bucket_names)}"
  bucket     = "${element(var.s3_bucket_names, count.index)}"
  depends_on = ["aws_s3_bucket.zbuckets"]

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "MYBUCKETPOLICY_{element(var.s3_bucket_names, count.index)}",
  "Statement": [
    {
      "Sid": "statementi_${element(var.s3_bucket_names, count.index)}",
      "Effect": "Allow",
      "Principal": {
      "AWS": ${jsonencode(aws_iam_user.bbl_s3_user.*.arn)}
       },
      "Action": "s3:*",
      "Resource": "arn:aws:s3:::${element(var.s3_bucket_names, count.index)}"
     }
  ]
}
POLICY
}

resource "aws_s3_bucket_policy" "bucketpol-replica" {
  count      = "${length(var.s3_bucket_names)}"
  bucket     = "${element(var.s3_bucket_names, count.index)}-replica"
  provider   = "aws.s3_replication_provider"
  depends_on = ["aws_s3_bucket.zbuckets"]

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "MYBUCKETPOLICY_{element(var.s3_bucket_names, count.index)}",
  "Statement": [
    {
      "Sid": "statementi_${element(var.s3_bucket_names, count.index)}",
      "Effect": "Allow",
      "Principal": {
      "AWS": ${jsonencode(aws_iam_user.bbl_s3_user.*.arn)}
       },
      "Action": "s3:*",
      "Resource": "arn:aws:s3:::${element(var.s3_bucket_names, count.index)}-replica"
     }
  ]
}
POLICY
}



resource "aws_iam_policy_attachment" "s3-replication" {
  name       = "${var.env_id}-s3-replication-policy-attachment"
  roles      = ["${aws_iam_role.s3-replication.name}"]
  policy_arn = "${aws_iam_policy.s3-replication.arn}"
}
resource "aws_s3_bucket" "zbuckets" {
  count      = "${length(var.s3_bucket_names)}"
  bucket     = "${element(var.s3_bucket_names, count.index)}"
  acl        = "private"
  depends_on = ["aws_s3_bucket.destinations"]

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        kms_master_key_id = "${aws_kms_key.kms_key_s3.arn}"
        sse_algorithm     = "aws:kms"
      }
    }
  }

  versioning {
    enabled = true
  }

  replication_configuration {
    role = "${aws_iam_role.s3-replication.arn}"

    rules {
      id     = "foobar"
      status = "Enabled"
      prefix = ""

      source_selection_criteria {
        sse_kms_encrypted_objects {
          enabled = "true"
        }
      }

      destination {
        bucket             = "arn:aws:s3:::${element(var.s3_bucket_names, count.index)}-replica"
        storage_class      = "STANDARD"
        replica_kms_key_id = "${aws_kms_key.kms_key_s3_rep.arn}"
      }
    }
  }
}

resource "aws_iam_user_policy" "bbl_s3_user_policy" {
  count      = "${length(var.s3_users)}"
  name       = "${element(var.s3_users, count.index)}"
  user       = "${element(var.s3_users, count.index)}"
  depends_on = ["aws_s3_bucket.zbuckets"]

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "s3:*",
            "Resource":
            ${jsonencode(concat(formatlist("%s/*",aws_s3_bucket.zbuckets.*.arn),formatlist("%s/*",aws_s3_bucket.destinations.*.arn)))}
        },
        {
            "Effect": "Allow",
            "Action": [
                "s3:ListAllMyBuckets",
                "s3:ListBucket",
                "s3:HeadBucket"
            ],
            "Resource": "*"

        },
        {
           "Effect": "Allow",
           "Action": "iam:ListRoles",
           "Resource": "*"
        }
    ]
}
EOF
}
