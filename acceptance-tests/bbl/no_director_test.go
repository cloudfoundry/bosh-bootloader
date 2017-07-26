package acceptance_test

import (
	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
		bbl.Destroy()
	})

	It("successfully standups up a no director infrastructure", func() {
		By("calling bbl up with the no-director flag", func() {
			bbl.Up(actors.GetIAAS(configuration), []string{"--name", bbl.PredefinedEnvID(), "--no-director"})
		})

		By("checking that no bosh director exists", func() {
			Expect(boshDirectorChecker.NetworkHasBOSHDirector(bbl.PredefinedEnvID())).To(BeFalse())
		})

		By("checking that director details are not printed", func() {
			directorUsername := bbl.DirectorUsername()
			Expect(directorUsername).To(Equal(""))
			directorPassword := bbl.DirectorPassword()
			Expect(directorPassword).To(Equal(""))
		})

		By("checking if bbl print-env prints the external ip", func() {
			stdout := bbl.PrintEnv()

			Expect(stdout).To(ContainSubstring("export BOSH_ENVIRONMENT="))
			Expect(stdout).NotTo(ContainSubstring("export BOSH_CLIENT="))
			Expect(stdout).NotTo(ContainSubstring("export BOSH_CLIENT_SECRET="))
			Expect(stdout).NotTo(ContainSubstring("export BOSH_CA_CERT="))
		})
	})
})
