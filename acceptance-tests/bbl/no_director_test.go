package acceptance_test

import (
	"time"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("no director test", func() {
	var (
		bbl                 actors.BBL
		state               acceptance.State
		configuration       acceptance.Config
		boshDirectorChecker actors.BOSHDirectorChecker
	)

	BeforeEach(func() {
		var err error
		configuration, err = acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "no-director-env")
		state = acceptance.NewState(configuration.StateFileDir)
		boshDirectorChecker = actors.NewBOSHDirectorChecker(configuration)
	})

	AfterEach(func() {
		session := bbl.Destroy()
		Eventually(session, 10*time.Minute).Should(gexec.Exit())
	})

	It("successfully standups up a no director infrastructure", func() {
		By("calling bbl up with the no-director flag", func() {
			session := bbl.Up("--name", bbl.PredefinedEnvID(), "--no-director")
			Eventually(session, 40*time.Minute).Should(gexec.Exit(0))
		})

		By("checking that no bosh director exists", func() {
			Expect(boshDirectorChecker.NetworkHasBOSHDirector(bbl.PredefinedEnvID())).To(BeFalse())
		})

		By("checking that director details are not printed", func() {
			Expect(bbl.DirectorUsername()).To(Equal(""))
			Expect(bbl.DirectorPassword()).To(Equal(""))
		})

		By("checking if bbl print-env prints the external ip", func() {
			stdout := bbl.PrintEnv()

			Expect(stdout).To(ContainSubstring("export BOSH_ENVIRONMENT="))
			Expect(stdout).NotTo(ContainSubstring("export BOSH_CLIENT="))
			Expect(stdout).NotTo(ContainSubstring("export BOSH_CLIENT_SECRET="))
			Expect(stdout).NotTo(ContainSubstring("export BOSH_CA_CERT="))
		})

		By("checking bbl up with no director is idempotent", func() {
			session := bbl.Up()
			Eventually(session, 40*time.Minute).Should(gexec.Exit(0))
		})

		By("checking that bosh director still does not exist", func() {
			Expect(boshDirectorChecker.NetworkHasBOSHDirector(bbl.PredefinedEnvID())).To(BeFalse())
		})
	})
})
