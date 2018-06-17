# Getting Started: AWS

This guide is a walkthrough for deploying a BOSH director with `bbl`
on AWS. Upon completion, you will have the following:

1. A BOSH director
1. A jumpbox
1. A set of randomly generated BOSH director credentials
1. A generated keypair allowing you to SSH into the BOSH director and
any instances BOSH deploys
1. A copy of the manifest the BOSH director was deployed with
1. A basic cloud config

## Creating an IAM user

In order for `bbl` to interact with AWS, an `IAM` user must be created.

This user will be issuing API requests to create the infrastructure such
as EC2 instances, load balancers, subnets, etc.

The user must have the following `policy`:

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
                "logs:*",
                "elasticloadbalancing:*",
                "cloudformation:*",
                "iam:*",
                "kms:*",
                "route53:*",
                "ec2:*"
            ],
            "Resource": "*"
        }
    ]
}
```

To create a user and associated policy with the AWS CLI run the 
following commands (policy text must be in your clipboard):

```
$ aws iam create-user --user-name "bbl-user"
$ aws iam put-user-policy --user-name "bbl-user" \
	--policy-name "bbl-policy" \
	--policy-document "$(pbpaste)"
$ aws iam create-access-key --user-name "bbl-user"
```

The `create-access-key` command will write an "access key id" and "secret 
access key" to the terminal. These values are important and should
be kept secret. In the next section `bbl` will use these commands to
create infrastructure on AWS.

## Pave Infrastructure, Create a Jumpbox, and Create a BOSH Director

`bbl` will create infrastructure and deploy a BOSH director with the
following command:

```
bbl up \
	--aws-access-key-id <INSERT ACCESS KEY ID> \
	--aws-secret-access-key <INSERT SECRET ACCESS KEY> \
	--aws-region us-west-1 \
	--iaas aws
```

The process takes around 5-8 minutes.

The bbl state directory contains all of the files that were used to
create your bosh director. This should be checked in to version control,
so that you have all the information necessary to later destroy or
update this environment at a later date.


## Next Steps

* [Target the BOSH Director](howto-target-bosh-director.md)
* [Deploy Cloud Foundry](cloudfoundry.md)
* [Deploy Concourse](concourse.md)