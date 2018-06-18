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
		bbl     actors.BBL
		boshcli actors.BOSHCLI

		directorAddress  string
		directorUsername string
		directorPassword string
		caCertPath       string

		stateDir   string
		iaas       string
		iaasHelper actors.IAASLBHelper
	)

	BeforeEach(func() {
		acceptance.SkipUnless("bbl-up")

		configuration, err := acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		iaas = configuration.IAAS
		iaasHelper = actors.NewIAASLBHelper(iaas, configuration)
		stateDir = configuration.StateFileDir

		bbl = actors.NewBBL(stateDir, pathToBBL, configuration, "up-env")
		boshcli = actors.NewBOSHCLI()
	})

	AfterEach(func() {
		By("destroying the director and the jumpbox", func() {
			session := bbl.Down()
			Eventually(session, 20*time.Minute).Should(gexec.Exit(0))
		})
	})

	It("bbl's up a new bosh director and jumpbox", func() {
		By("cleaning up any leftovers", func() {
			session := bbl.CleanupLeftovers(bbl.PredefinedEnvID())
			Eventually(session, 10*time.Minute).Should(gexec.Exit())
		})

		args := []string{
			"--name", bbl.PredefinedEnvID(),
		}
		args = append(args, iaasHelper.GetLBArgs()...)
		session := bbl.Up(args...)
		Eventually(session, 60*time.Minute).Should(gexec.Exit(0))

		By("exporting bosh environment variables", func() {
			bbl.ExportBoshAllProxy()
		})

		By("checking if the bosh director exists via the bosh cli", func() {
			directorAddress = bbl.DirectorAddress()
			directorUsername = bbl.DirectorUsername()
			directorPassword = bbl.DirectorPassword()
			caCertPath = bbl.SaveDirectorCA()

			directorExists := func() bool {
				exists, err := boshcli.DirectorExists(directorAddress, directorUsername, directorPassword, caCertPath)
				if err != nil {
					fmt.Println(string(err.(*exec.ExitError).Stderr))
				}
				return exists
			}
			Eventually(directorExists, "1m", "10s").Should(BeTrue())
		})

		By("verifying that vm extensions were added to the cloud config", func() {
			cloudConfig, err := boshcli.CloudConfig(directorAddress, caCertPath, directorUsername, directorPassword)
			Expect(err).NotTo(HaveOccurred())

			vmExtensions := acceptance.VmExtensionNames(cloudConfig)
			iaasHelper.VerifyCloudConfigExtensions(vmExtensions)
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

			session := bbl.Rotate()
			Eventually(session, 40*time.Minute).Should(gexec.Exit(0))

			rotatedKey := bbl.SSHKey()
			Expect(rotatedKey).NotTo(BeEmpty())
			Expect(rotatedKey).NotTo(Equal(sshKey))
		})

		By("checking bbl up is idempotent", func() {
			session := bbl.Up()
			Eventually(session, 40*time.Minute).Should(gexec.Exit(0))
		})

		By("confirming that the load balancers exist", func() {
			iaasHelper.ConfirmLBsExist(bbl.PredefinedEnvID())
		})

		By("verifying the bbl lbs output", func() {
			stdout := bbl.Lbs()
			iaasHelper.VerifyBblLBOutput(stdout)
		})

		By("deleting lbs", func() {
			session := bbl.Plan("--name", bbl.PredefinedEnvID())
			Eventually(session, 1*time.Minute).Should(gexec.Exit(0))

			session = bbl.Up()
			Eventually(session, 60*time.Minute).Should(gexec.Exit(0))
		})

		By("confirming that the load balancers no longer exist", func() {
			iaasHelper.ConfirmNoLBsExist(bbl.PredefinedEnvID())
		})
	})
})
