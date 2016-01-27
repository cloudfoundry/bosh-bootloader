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

Currently the `bbl` command is a no-op, but you can invoke it with:

```
$ bbl
```

test
