package commands

import (
	"fmt"
	"io"
	"strings"

	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

const (
	GLOBAL_USAGE = `
Usage:
  bbl [GLOBAL OPTIONS] %s [OPTIONS]

Global Options:
  --help    [-h] "print usage"
  --version [-v] "print version"

  --state-dir             "Directory that stores the state.json"
%s
`
	COMMAND_USAGE = `
[%s command options]
  %s`
)

const USAGE = `
Commands:
  destroy [--no-confirm]                                                                                               "tears down a BOSH Director environment on AWS"
  director-address                                                                                                     "prints the BOSH director address"
  director-username                                                                                                    "prints the BOSH director username"
  director-password                                                                                                    "prints the BOSH director password"
  bosh-ca-cert                                                                                                         "prints the BOSH director CA certificate"
  env-id                                                                                                               "prints the environment ID"
  help                                                                                                                 "prints usage"
  lbs                                                                                                                  "prints any attached load balancers"
  ssh-key                                                                                                              "prints the ssh private key"
  create-lbs --type=<concourse,cf> --cert=<path> --key=<path> [--chain=<path>] [--skip-if-exists]                      "attaches a load balancer with the supplied certificate, key, and optional chain"
  update-lbs --cert=<path> --key=<path> [--chain=<path>] [--skip-if-missing]                                           "updates a load balancer with the supplied certificate, key, and optional chain"
  delete-lbs [--skip-if-missing]                                                                                       "deletes the attached load balancer"
  up --aws-access-key-id <aws_access_key_id> --aws-secret-access-key <aws_secret_access_key> --aws-region <aws_region> "deploys a BOSH Director on AWS"
  version                                                                                                              "prints version"`

type Usage struct {
	stdout io.Writer
}

func NewUsage(stdout io.Writer) Usage {
	return Usage{stdout}
}

func (u Usage) Execute(subcommandFlags []string, state storage.State) error {
	u.Print()
	return nil
}

func (u Usage) Print() {
	content := fmt.Sprintf(GLOBAL_USAGE, "COMMAND", USAGE)
	fmt.Fprint(u.stdout, strings.TrimLeft(content, "\n"))
}

func (u Usage) PrintCommandUsage(command, message string) {
	commandUsage := fmt.Sprintf(COMMAND_USAGE, command, message)
	content := fmt.Sprintf(GLOBAL_USAGE, command, commandUsage)
	fmt.Fprint(u.stdout, strings.TrimLeft(content, "\n"))
}
