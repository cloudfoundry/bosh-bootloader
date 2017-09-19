package acceptance_test

import (
	"fmt"
	"time"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
)

var _ = Describe("up test", func() {
	var (
		azure   actors.Azure
		bbl     actors.BBL
		boshcli actors.BOSHCLI
	)

	BeforeEach(func() {
		config, err := acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		azure = actors.NewAzure(config)
		bbl = actors.NewBBL(config.StateFileDir, pathToBBL, config, "azure-env")
		boshcli = actors.NewBOSHCLI()
	})

	AfterEach(func() {
		session := bbl.Down()
		Eventually(session, 10*time.Minute).Should(gexec.Exit())
	})

	It("creates the director", func() {
		session := bbl.Up("--name", bbl.PredefinedEnvID())
		Eventually(session, 40*time.Minute).Should(gexec.Exit(0))

		exists, err := azure.GetResourceGroup(fmt.Sprintf("%s-bosh", bbl.PredefinedEnvID()))
		Expect(err).NotTo(HaveOccurred())
		Expect(exists).To(BeTrue())

		directorAddress := bbl.DirectorAddress()
		caCertPath := bbl.SaveDirectorCA()
		By("checking if the bosh director exists", func() {
			exists, err := boshcli.DirectorExists(directorAddress, caCertPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		By("checking that the cloud config exists", func() {
			directorUsername := bbl.DirectorUsername()
			directorPassword := bbl.DirectorPassword()

			cloudConfig, err := boshcli.CloudConfig(directorAddress, caCertPath, directorUsername, directorPassword)
			Expect(err).NotTo(HaveOccurred())
			Expect(cloudConfig).NotTo(BeEmpty())
		})

		By("checking if bbl print-env prints the bosh environment variables", func() {
			stdout := bbl.PrintEnv()

			Expect(stdout).To(ContainSubstring("export BOSH_ENVIRONMENT="))
			Expect(stdout).To(ContainSubstring("export BOSH_CLIENT="))
			Expect(stdout).To(ContainSubstring("export BOSH_CLIENT_SECRET="))
			Expect(stdout).To(ContainSubstring("export BOSH_CA_CERT="))
		})

		By("checking bbl up with director is idempotent", func() {
			session := bbl.Up()
			Eventually(session, 40*time.Minute).Should(gexec.Exit(0))
		})

		By("destroying the director", func() {
			session := bbl.Down()
			Eventually(session, 10*time.Minute).Should(gexec.Exit(0))
		})
	})
})
