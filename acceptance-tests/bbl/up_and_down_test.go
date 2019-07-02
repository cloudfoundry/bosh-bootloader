package acceptance_test

import (
	"fmt"
	"os/exec"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("up_and_down", func() {
	var (
		bbl     actors.BBL
		boshcli actors.BOSHCLI

		directorAddress  string
		directorUsername string
		directorPassword string
		caCertPath       string

		stateDir    string
		iaas        string
		stemcellURL string
		iaasHelper  actors.IAASLBHelper
	)

	BeforeEach(func() {
		acceptance.SkipUnless("bbl-up")

		configuration, err := acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		iaas = configuration.IAAS
		stemcellURL = configuration.StemcellURL
		iaasHelper = actors.NewIAASLBHelper(iaas, configuration)
		stateDir = configuration.StateFileDir

		bbl = actors.NewBBL(stateDir, pathToBBL, configuration, "up-env", false)
		boshcli = actors.NewBOSHCLI()
	})

	AfterEach(func() {
		By("ensure the director and the jumpbox are destroyed", func() {
			session := bbl.Down()
			Eventually(session, bblDownTimeout).Should(gexec.Exit(0))
		})
	})

	It("bbl's up a new bosh director and jumpbox", func() {
		By("cleaning up any leftovers", func() {
			session := bbl.CleanupLeftovers(bbl.PredefinedEnvID())
			Eventually(session, bblLeftoversTimeout).Should(gexec.Exit())
		})

		args := []string{
			"--name", bbl.PredefinedEnvID(),
		}
		args = append(args, iaasHelper.GetLBArgs()...)
		session := bbl.Up(args...)
		Eventually(session, bblUpTimeout).Should(gexec.Exit(0))

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

		By("verifying that the bosh dns runtime config was set", func() {
			_, err := boshcli.RuntimeConfig(directorAddress, caCertPath, directorUsername, directorPassword, "dns")
			Expect(err).NotTo(HaveOccurred())
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
			Eventually(session, bblRotateTimeout).Should(gexec.Exit(0))

			rotatedKey := bbl.SSHKey()
			Expect(rotatedKey).NotTo(BeEmpty())
			Expect(rotatedKey).NotTo(Equal(sshKey))
		})

		By("resetting the correct creds", func() {
			bbl.ExportBoshAllProxy()
			directorAddress = bbl.DirectorAddress()
			directorUsername = bbl.DirectorUsername()
			directorPassword = bbl.DirectorPassword()
			caCertPath = bbl.SaveDirectorCA()
		})

		By("checking bbl up is idempotent", func() {
			session := bbl.Up()
			Eventually(session, bblUpTimeout).Should(gexec.Exit(0))
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
			Eventually(session, bblPlanTimeout).Should(gexec.Exit(0))

			session = bbl.Up()
			Eventually(session, bblUpTimeout).Should(gexec.Exit(0))
		})

		By("confirming that the load balancers no longer exist", func() {
			iaasHelper.ConfirmNoLBsExist(bbl.PredefinedEnvID())
		})

		if stemcellURL != "" {
			When("stemcells are uploaded and bbl down is called", func() {
				err := boshcli.UploadStemcell(directorAddress, caCertPath, directorUsername, directorPassword, stemcellURL)
				Expect(err).NotTo(HaveOccurred())

				stemcellIDs, err := boshcli.Stemcells(directorAddress, caCertPath, directorUsername, directorPassword)
				Expect(err).NotTo(HaveOccurred())

				By("destroy director and the jumpbox", func() {
					session := bbl.Down()
					Eventually(session, bblDownTimeout).Should(gexec.Exit(0))
				})

				By("removes created stemcells from iaas", func() {
					iaasHelper.ConfirmNoStemcellsExist(stemcellIDs)
				})
			})
		} else {
			By("destroy director and the jumpbox", func() {
				session := bbl.Down()
				Eventually(session, bblDownTimeout).Should(gexec.Exit(0))
			})
		}

	})
})
