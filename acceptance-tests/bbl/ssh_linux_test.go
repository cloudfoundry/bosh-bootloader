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

var _ = Describe("ssh", func() {
	var (
		bbl actors.BBL

		stateDir   string
		iaas       string
		iaasHelper actors.IAASLBHelper
	)

	addKnownHost := func() {
		script := `mkdir -p ~/.ssh/; \
			ssh-keyscan -H -trsa %s >> ~/.ssh/known_hosts`
		_, err := exec.Command("sh", "-c",
			fmt.Sprintf(script, bbl.JumpboxAddress()),
		).Output()
		Expect(err).NotTo(HaveOccurred())
	}

	removeKnownHost := func() {
		exec.Command("ssh-keygen", "-R", bbl.JumpboxAddress()).Output()
	}

	BeforeEach(func() {
		acceptance.SkipUnless("ssh")

		configuration, err := acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		iaas = configuration.IAAS
		iaasHelper = actors.NewIAASLBHelper(iaas, configuration)
		stateDir = configuration.StateFileDir

		bbl = actors.NewBBL(stateDir, pathToBBL, configuration, "ssh-env", false)
	})

	AfterEach(func() {
		By("destroying the director and the jumpbox", func() {
			session := bbl.Down()
			Eventually(session, bblDownTimeout).Should(gexec.Exit(0))
		})

		removeKnownHost()
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

		By("noninteractively preverifying the jumpbox's public key", func() {
			addKnownHost()
		})

		By("checking to see if we can ssh to the jumpbox", func() {
			bbl.JumpboxSSH(GinkgoWriter)
		})

		By("verifying we can ssh to the director", func() {
			bbl.DirectorSSH(GinkgoWriter)
		})

		By("rotating the jumpbox's ssh key", func() {
			sshKey := bbl.SSHKey()
			Expect(sshKey).NotTo(BeEmpty())

			session := bbl.Rotate()
			Eventually(session, bblRotateTimeout).Should(gexec.Exit(0))

			rotatedKey := bbl.SSHKey()
			Expect(rotatedKey).NotTo(BeEmpty())
			Expect(rotatedKey).NotTo(Equal(sshKey))

			removeKnownHost()
			addKnownHost()
		})

		By("checking to see if we can still ssh to the jumpbox", func() {
			bbl.JumpboxSSH(GinkgoWriter)
		})

		By("checking to see if we can still ssh to the director", func() {
			bbl.DirectorSSH(GinkgoWriter)
		})
	})
})
