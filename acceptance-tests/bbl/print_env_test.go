package acceptance_test

import (
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
		stateDir = configuration.StateFileDir

		bbl = actors.NewBBL(stateDir, pathToBBL, configuration, "fixture-state", true)

		// 	session := bbl.Plan("--name", bbl.PredefinedEnvID())
		// 	Eventually(session, 5*time.Minute).Should(gexec.Exit(0))
		// TODO: always upload the same bbl state so we don't have fixture rot
	})

	FIt("can print a bbl environment stored in the THE CLOUD", func() {
		stdout := bbl.PrintEnv()
		Expect(stdout).To(ContainSubstring("export BOSH_CLIENT=admin"))
		Expect(stdout).To(ContainSubstring("export BOSH_CLIENT_SECRET=some-password"))
		Expect(stdout).To(ContainSubstring("export BOSH_ENVIRONMENT=https://10.0.0.6:25555"))
		Expect(stdout).To(ContainSubstring("export BOSH_CA_CERT='-----BEGIN CERTIFICATE-----\ndirector-ca-cert\n-----END CERTIFICATE-----"))
		Expect(stdout).To(ContainSubstring("export JUMPBOX_PRIVATE_KEY="))
		Expect(stdout).To(ContainSubstring("export BOSH_ALL_PROXY=ssh+socks5://jumpbox@35.185.60.196:22?private-key="))
		Expect(stdout).To(ContainSubstring("bosh_jumpbox_private.key"))
	})
})
