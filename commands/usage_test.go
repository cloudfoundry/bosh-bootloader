package commands_test

import (
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Usage", func() {
	var (
		usage  commands.Usage
		logger *fakes.Logger
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}

		usage = commands.NewUsage(logger)
	})

	Describe("CheckFastFails", func() {
		It("returns no error", func() {
			err := usage.CheckFastFails([]string{}, storage.State{})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Execute", func() {
		It("prints out the usage information", func() {
			err := usage.Execute([]string{}, storage.State{})
			Expect(err).NotTo(HaveOccurred())
			Expect(logger.PrintlnCall.Receives.Message).To(Equal(strings.TrimLeft(`
Usage:
  bbl [GLOBAL OPTIONS] COMMAND [OPTIONS]

Global Options:
  --help      [-h]       Prints usage
  --state-dir            Directory containing bbl-state.json
  --debug                Prints debugging output
  --version              Prints version

Commands:
  help                    Prints usage
  version                 Prints version
  up                      Deploys BOSH director on an IAAS
  destroy                 Tears down BOSH director infrastructure
  lbs                     Prints attached load balancer(s)
  create-lbs              Attaches load balancer(s)
  update-lbs              Updates load balancer(s)
  delete-lbs              Deletes attached load balancer(s)
  rotate                  Rotates SSH key for the jumpbox user
  bosh-deployment-vars    Prints required variables for BOSH deployment
  jumpbox-deployment-vars Prints required variables for jumpbox deployment
  cloud-config            Prints suggested cloud configuration for BOSH environment
  jumpbox-address         Prints BOSH jumpbox address
  director-address        Prints BOSH director address
  director-username       Prints BOSH director username
  director-password       Prints BOSH director password
  director-ca-cert        Prints BOSH director CA certificate
  env-id                  Prints environment ID
  latest-error            Prints the output from the latest call to terraform
  print-env               Prints BOSH friendly environment variables
  ssh-key                 Prints SSH private key

  Use "bbl [command] --help" for more information about a command.
`, "\n")))
		})
	})

	Describe("PrintCommandUsage", func() {
		It("prints the usage for given command", func() {
			usage.PrintCommandUsage("my-command", "some message")

			Expect(logger.PrintlnCall.Receives.Message).To(Equal(strings.TrimLeft(`Usage:
  bbl [GLOBAL OPTIONS] my-command [OPTIONS]

Global Options:
  --help      [-h]       Prints usage
  --state-dir            Directory containing bbl-state.json
  --debug                Prints debugging output
  --version              Prints version

[my-command command options]
  some message
`, "\n")))
		})
	})
})
