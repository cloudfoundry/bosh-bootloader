package commands_test

import (
	"bytes"
	"strings"

	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Usage", func() {
	var (
		usage  commands.Usage
		stdout *bytes.Buffer
	)

	BeforeEach(func() {
		stdout = bytes.NewBuffer([]byte{})
		usage = commands.NewUsage(stdout)
	})

	Describe("Execute", func() {
		It("prints out the usage information", func() {
			err := usage.Execute([]string{}, storage.State{})
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout.String()).To(Equal(strings.TrimLeft(`
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
  env-id                                                                                                      "prints the environment ID"
  help                                                                                                        "prints usage"
  lbs                                                                                                         "prints any attached load balancers"
  ssh-key                                                                                                     "prints the ssh private key"
  create-lbs --type=<concourse,cf> --cert=<path> --key=<path> [--chain=<path>] [--skip-if-exists]             "attaches a load balancer with the supplied certificate, key, and optional chain"
  update-lbs --cert=<path> --key=<path> [--chain=<path>] [--skip-if-missing]                                  "updates a load balancer with the supplied certificate, key, and optional chain"
  delete-lbs [--skip-if-missing]                                                                              "deletes the attached load balancer"
  up                                                                                                          "deploys a BOSH Director on AWS"
  version                                                                                                     "prints version"
`, "\n")))
		})
	})
})
