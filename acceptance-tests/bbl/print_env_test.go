package acceptance_test

import (
	"fmt"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("plan", func() {
	var (
		bbl      actors.BBL
		stateDir string
		iaas     string
	)

	BeforeEach(func() {
		acceptance.SkipUnless("plan")

		configuration, err := acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		iaas = configuration.IAAS
		if iaas != "aws" && iaas != "gcp" {
			Skip(fmt.Sprintf("won't run remote bbl state tests on IAAS %q ", iaas))
		}

		stateDir = configuration.StateFileDir

		configuration.BBLStateBucket = "bbl-acceptance-test-states"

		stateFileName := fmt.Sprintf("fixture-state-%s", iaas)
		bbl = actors.NewBBL(stateDir, pathToBBL, configuration, stateFileName, true)

		// TODO: always upload the same bbl state so we don't have fixture rot
	})

	It("can print a bbl environment stored in the THE CLOUD", func() {
		stdout := bbl.PrintEnv()
		Expect(stdout).To(ContainSubstring("export BOSH_CLIENT=admin"))
		Expect(stdout).To(ContainSubstring("export BOSH_CLIENT_SECRET=some-password"))
		Expect(stdout).To(ContainSubstring("export BOSH_ENVIRONMENT=https://10.0.0.6:25555"))
		Expect(stdout).To(ContainSubstring("export BOSH_CA_CERT='-----BEGIN CERTIFICATE-----\ndirector-ca-cert\n-----END CERTIFICATE-----"))
		Expect(stdout).To(ContainSubstring("export JUMPBOX_PRIVATE_KEY="))
		Expect(stdout).To(ContainSubstring("export BOSH_ALL_PROXY=ssh+socks5://jumpbox@10.0.1.5:22?private-key="))
		Expect(stdout).To(ContainSubstring("bosh_jumpbox_private.key"))
	})
})
