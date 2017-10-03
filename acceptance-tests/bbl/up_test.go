package acceptance_test

import (
	"fmt"
	"os/exec"
	"time"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("up", func() {
	var (
		bbl             actors.BBL
		boshcli         actors.BOSHCLI
		directorAddress string
		caCertPath      string
		sshSession      *gexec.Session

		boshDirectorChecker actors.BOSHDirectorChecker
	)

	BeforeEach(func() {
		configuration, err := acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "up-env")
		boshcli = actors.NewBOSHCLI()

		boshDirectorChecker = actors.NewBOSHDirectorChecker(configuration)
	})

	AfterEach(func() {
		fmt.Println("********************************************************************************")
		fmt.Println("Entering AfterEach...")
		sshSession.Interrupt()
		fmt.Println("Called sshSession.Interrupt...")
		Eventually(sshSession, "5s").Should(gexec.Exit())
		fmt.Println("sshSession exited...")
		session := bbl.Down()
		fmt.Println("Called bbl.Down()...")
		Eventually(session, 10*time.Minute).Should(gexec.Exit())
		fmt.Println("Exiting AfterEach...")
		fmt.Println("********************************************************************************")
	})

	It("bbl's up a new bosh director and jumpbox", func() {
		acceptance.SkipUnless("bbl-up")
		session := bbl.Up("--name", bbl.PredefinedEnvID())
		Eventually(session, 40*time.Minute).Should(gexec.Exit(0))

		By("creating an ssh tunnel to the director in print-env", func() {
			sshSession = bbl.StartSSHTunnel()
		})

		By("checking if the bosh director exists", func() {
			directorAddress = bbl.DirectorAddress()
			caCertPath = bbl.SaveDirectorCA()

			directorExists := func() bool {
				exists, err := boshcli.DirectorExists(directorAddress, caCertPath)
				if err != nil {
					fmt.Println(string(err.(*exec.ExitError).Stderr))
				}
				return exists
			}
			Eventually(directorExists, "1m", "10s").Should(BeTrue())
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

		By("rotating the jumpbox's ssh key", func() {
			sshKey := bbl.SSHKey()
			Expect(sshKey).NotTo(BeEmpty())

			session = bbl.Rotate()
			Eventually(session, 40*time.Minute).Should(gexec.Exit(0))

			rotatedKey := bbl.SSHKey()
			Expect(rotatedKey).NotTo(BeEmpty())
			Expect(rotatedKey).NotTo(Equal(sshKey))
		})

		By("checking bbl up is idempotent", func() {
			session := bbl.Up()
			Eventually(session, 40*time.Minute).Should(gexec.Exit(0))
		})

		By("destroying the director and the jumpbox", func() {
			session := bbl.Down()
			Eventually(session, 10*time.Minute).Should(gexec.Exit(0))
		})
	})
})
