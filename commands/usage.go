package commands

import (
	"fmt"
	"io"
	"strings"

	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
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
  destroy [--no-confirm]                                                                                      "tears down a BOSH Director environment on AWS"
  director-address                                                                                            "prints the BOSH director address"
  director-username                                                                                           "prints the BOSH director username"
  director-password                                                                                           "prints the BOSH director password"
  bosh-ca-cert                                                                                                "prints the BOSH director CA certificate"
  help                                                                                                        "prints usage"
  lbs                                                                                                         "prints any attached load balancers"
  ssh-key                                                                                                     "prints the ssh private key"
  unsupported-create-lbs --type=<concourse,cf> --cert=<path> --key=<path> [--chain=<path>] [--skip-if-exists] "attaches a load balancer with the supplied certificate, key, and optional chain"
  unsupported-update-lbs --cert=<path> --key=<path> [--chain=<path>] [--skip-if-missing]                      "updates a load balancer with the supplied certificate, key, and optional chain"
  unsupported-delete-lbs [--skip-if-missing]                                                                  "deletes the attached load balancer"
  unsupported-deploy-bosh-on-aws-for-concourse                                                                "deploys a BOSH Director on AWS"
  version                                                                                                     "prints version"
`

type Usage struct {
	stdout io.Writer
}

func NewUsage(stdout io.Writer) Usage {
	return Usage{stdout}
}

func (u Usage) Execute(subcommandFlags []string, state storage.State) (storage.State, error) {
	u.Print()
	return state, nil
}

func (u Usage) Print() {
	fmt.Fprint(u.stdout, strings.TrimLeft(USAGE, "\n"))
}
