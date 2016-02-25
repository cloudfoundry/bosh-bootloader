package commands

import (
	"fmt"
	"io"
	"strings"

	"github.com/pivotal-cf-experimental/bosh-bootloader/state"
)

const USAGE = `
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
  help                                     "print usage"
  version                                  "print version"
  unsupported-print-concourse-aws-template "print a concourse aws template"
  unsupported-create-bosh-aws-keypair      "create and upload a keypair to AWS"
  unsupported-provision-aws-for-concourse  "create a new concourse stack on AWS"
`

type Usage struct {
	stdout io.Writer
}

func NewUsage(stdout io.Writer) Usage {
	return Usage{stdout}
}

func (u Usage) Execute(globalFlags GlobalFlags, s state.State) (state.State, error) {
	u.Print()
	return s, nil
}

func (u Usage) Print() {
	fmt.Fprint(u.stdout, strings.TrimSpace(USAGE))
}
