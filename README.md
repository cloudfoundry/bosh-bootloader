# bosh-bootloader
---

This is a command line utility for standing up a CloudFoundry or Concourse installation 
on an IAAS of your choice. This CLI is currently under heavy development, and the
initial goal is to support bootstrapping a CloudFoundry installation on AWS.

* [CI](https://mega.ci.cf-app.com/pipelines/bosh-bootloader)
* [Tracker](https://www.pivotaltracker.com/n/projects/1488988)

## Prerequisites

The following should be installed on your local machine
- Golang (>= 1.5)

If using homebrew, these can be installed with:

```
brew install go
```

## Installation

```
go get github.com/pivotal-cf-experimental/bosh-bootloader/...
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
  help                                          "print usage"
  version                                       "print version"
  unsupported-deploy-bosh-on-aws-for-concourse  "deploys a BOSH Director on AWS"
  destroy                                       "tears down a BOSH Director environment on AWS"
```
