package commands

import (
	"fmt"
	"io"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

const (
	HelpCommand = "help"

	UsageHeader = `
Usage:
  bbl [GLOBAL OPTIONS] %s [OPTIONS]

Global Options:
  --help      [-h]       Prints usage
  --state-dir            Directory containing bbl-state.json
  --debug                Prints debugging output
%s
`
	CommandUsage = `
[%s command options]
  %s`
)

const GlobalUsage = `
Commands:
  bosh-ca-cert           Prints BOSH director CA certificate
  bosh-deployment-vars   Prints required variables for BOSH deployment
  cloud-config           Prints suggested cloud configuration for BOSH environment
  create-lbs             Attaches load balancer(s)
  delete-lbs             Deletes attached load balancer(s)
  destroy                Tears down BOSH director infrastructure
  director-address       Prints BOSH director address
  director-username      Prints BOSH director username
  director-password      Prints BOSH director password
  director-ca-cert       Prints BOSH director CA certificate
  env-id                 Prints environment ID
  print-env              Prints BOSH friendly environment variables
  help                   Prints usage
  lbs                    Prints attached load balancer(s)
  ssh-key                Prints SSH private key
  up                     Deploys BOSH director on AWS
  update-lbs             Updates load balancer(s)
  version                Prints version

  Use "bbl [command] --help" for more information about a command.`

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
	content := fmt.Sprintf(UsageHeader, "COMMAND", GlobalUsage)
	fmt.Fprint(u.stdout, strings.TrimLeft(content, "\n"))
}

func (u Usage) PrintCommandUsage(command, message string) {
	commandUsage := fmt.Sprintf(CommandUsage, command, message)
	content := fmt.Sprintf(UsageHeader, command, commandUsage)
	fmt.Fprint(u.stdout, strings.TrimLeft(content, "\n"))
}
