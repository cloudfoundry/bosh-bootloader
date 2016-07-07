# bosh-bootloader
---

This is a command line utility for standing up a Concourse or [CloudFoundry installation](README-cloudfoundry.md)
on an IAAS of your choice. This CLI is currently under heavy development, and the
initial goal is to support bootstrapping a CloudFoundry installation on AWS.

* [CI](https://mega.ci.cf-app.com/pipelines/bosh-bootloader)
* [Tracker](https://www.pivotaltracker.com/n/projects/1488988)

## Prerequisites

### Install Dependencies

The following should be installed on your local machine
- Golang >= 1.5 (install with `brew install go` for Mac OS X, download the binary or package for Linux)
- bosh-init ([installation instructions](http://bosh.io/docs/install-bosh-init.html))
- Ruby 2+ (check with `ruby -v`) and make sure you did `gem install bosh_cli`

### Install bosh-bootloader

bosh-bootloader can be installed with go get:

```
go get github.com/pivotal-cf-experimental/bosh-bootloader/bbl
```

### Configure AWS

The AWS IAM user that is provided to bbl will need the following policy:

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ec2:*",
                "cloudformation:*",
                "elasticloadbalancing:*",
                "iam:*"
            ],
            "Resource": [
                "*"
            ]
        }
    ]
}
```

## Usage

The `bbl` command can be invoked on the command line and will display its usage.

```
$ bbl
Usage:
  bbl [GLOBAL OPTIONS] COMMAND [OPTIONS]

Global Options:
  --help    [-h] "print usage"
  --version [-v] "print version"

  --aws-access-key-id     "AWS AccessKeyID value"
  --aws-secret-access-key "AWS SecretAccessKey value"
  --aws-region            "AWS Region"
  --state-dir             "Directory that stores the state.json"

Commands:
  destroy [--no-confirm]                                                                                      "tears down a BOSH Director environment on AWS"
  director-address                                                                                            "print the BOSH director address"
  director-password                                                                                           "print the BOSH director password"
  director-username                                                                                           "print the BOSH director username"
  help                                                                                                        "print usage"
  lbs                                                                                                         "prints any attached load balancers"
  ssh-key                                                                                                     "print the ssh private key"
  unsupported-create-lbs --type=<concourse,cf> --cert=<path> --key=<path> [--chain=<path>] [--skip-if-exists] "attaches a load balancer with the supplied certificate, key, and optional chain"
  unsupported-update-lbs --cert=<path> --key=<path> [--chain=<path>] [--skip-if-missing]                      "updates a load balancer with the supplied certificate, key, and optional chain"
  unsupported-delete-lbs [--skip-if-missing]                                                                  "deletes the attached load balancer"
  unsupported-deploy-bosh-on-aws-for-concourse                                                                "deploys a BOSH Director on AWS"
  version                                                                                                     "print version"
```
