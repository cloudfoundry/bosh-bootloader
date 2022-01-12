variable "bosh_iam_instance_profile" {
  default = ""
}

locals {
  iamProfileProvided = var.bosh_iam_instance_profile == "" ? false : true
  iamProfileCount    = var.bosh_iam_instance_profile == "" ? 0 : 1
}

data "aws_iam_instance_profile" "bosh" {
  name = var.bosh_iam_instance_profile

  count = local.iamProfileCount
}

resource "aws_iam_role" "bosh" {
  name = "${var.env_id}_bosh_role"
  path = "/"

  count = 1 - local.iamProfileCount

  lifecycle {
    create_before_destroy = true
  }

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "bosh" {
  name = "${var.env_id}_bosh_policy"
  path = "/"

  count = 1 - local.iamProfileCount

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:AssociateAddress",
        "ec2:AttachVolume",
        "ec2:CopyImage",
        "ec2:CreateVolume",
        "ec2:DeleteSnapshot",
        "ec2:DeleteVolume",
        "ec2:DescribeAddresses",
        "ec2:DescribeAvailabilityZones",
        "ec2:DescribeImages",
        "ec2:DescribeInstances",
        "ec2:DescribeRegions",
        "ec2:DescribeSecurityGroups",
        "ec2:DescribeSnapshots",
        "ec2:DescribeSubnets",
        "ec2:DescribeVolumes",
        "ec2:DetachVolume",
        "ec2:CreateSnapshot",
        "ec2:CreateTags",
        "ec2:ModifyInstanceAttribute",
        "ec2:RunInstances",
        "ec2:TerminateInstances",
        "ec2:RegisterImage",
        "ec2:DeregisterImage",
        "ec2:CancelSpotInstanceRequests",
        "ec2:DescribeSpotInstanceRequests",
        "ec2:RequestSpotInstances",
        "ec2:CreateRoute",
        "ec2:DescribeRouteTables",
        "ec2:ReplaceRoute"
	  ],
	  "Effect": "Allow",
	  "Resource": "*"
    },
	{
	  "Action": [
            "iam:PassRole",
            "iam:CreateServiceLinkedRole"
	  ],
	  "Effect": "Allow",
	  "Resource": "*"
	},
	{
	  "Action": [
	    "elasticloadbalancing:*"
	  ],
	  "Effect": "Allow",
	  "Resource": "*"
	},
        {
            "Effect": "Allow",
            "Action": [
                "kms:ReEncrypt*",
                "kms:GenerateDataKey*",
                "kms:CreateGrant",
                "kms:DescribeKey*"
            ],
            "Resource": [
                "*"
            ]
        }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "bosh" {
  role       = "${var.env_id}_bosh_role"
  policy_arn = aws_iam_policy.bosh[0].arn

  count = 1 - local.iamProfileCount
}

resource "aws_iam_instance_profile" "bosh" {
  name = "${var.env_id}-bosh"
  role = aws_iam_role.bosh[0].name

  count = 1 - local.iamProfileCount

  lifecycle {
    ignore_changes = [name]
  }
}

resource "aws_flow_log" "bbl" {
  log_destination = aws_cloudwatch_log_group.bbl.arn
  iam_role_arn    = aws_iam_role.flow_logs.arn
  vpc_id          = local.vpc_id
  traffic_type    = "REJECT"
}

resource "aws_cloudwatch_log_group" "bbl" {
  name_prefix = "${var.short_env_id}-log-group"

  tags = {
    Name = "${var.env_id}"
  }
}

resource "aws_iam_role" "flow_logs" {
  name = "${var.env_id}-flow-logs-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "vpc-flow-logs.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "flow_logs" {
  name = "${var.env_id}-flow-logs-policy"
  role = aws_iam_role.flow_logs.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents",
        "logs:DescribeLogGroups",
        "logs:DescribeLogStreams"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

output "iam_instance_profile" {
  value = local.iamProfileProvided ? join("", data.aws_iam_instance_profile.bosh.*.name) : join("", aws_iam_instance_profile.bosh.*.name)
}
