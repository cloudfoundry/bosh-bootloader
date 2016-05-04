# bosh-bootloader
---

This is a command line utility for standing up a CloudFoundry or Concourse installation 
on an IAAS of your choice. This CLI is currently under heavy development, and the
initial goal is to support bootstrapping a CloudFoundry installation on AWS.

* [CI](https://mega.ci.cf-app.com/pipelines/bosh-bootloader)
* [Tracker](https://www.pivotaltracker.com/n/projects/1488988)

## Prerequisites

### Install Dependencies

The following should be installed on your local machine
- Golang (>= 1.5)

If using homebrew, these can be installed with:

```
brew install go
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

## Installation

bbl requires bosh-init. Instructions on how to install bosh-init [can be found here](http://bosh.io/docs/install-bosh-init.html)

```
go get github.com/pivotal-cf-experimental/bosh-bootloader/bbl
```

## Usage

The `bbl` command can be invoked on the command line and will display it's usage.

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
  destroy [--no-confirm]                                                      "tears down a BOSH Director environment on AWS"
  director-address                                                            "print the BOSH director address"
  director-password                                                           "print the BOSH director password"
  director-username                                                           "print the BOSH director username"
  help                                                                        "print usage"
  ssh-key                                                                     "print the ssh private key"
  unsupported-deploy-bosh-on-aws-for-concourse [--lb-type=concourse,cf,none]  "deploys a BOSH Director on AWS"
  version                                                                     "print version"
```
