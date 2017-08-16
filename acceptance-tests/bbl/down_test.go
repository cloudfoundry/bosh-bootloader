package acceptance_test

import (
	"time"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("down", func() {
	var (
		bbl           actors.BBL
		boshcli       actors.BOSHCLI
		configuration acceptance.Config
	)

	BeforeEach(func() {
		var err error
		configuration, err = acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "down-env")
		boshcli = actors.NewBOSHCLI()
	})

	AfterEach(func() {
		session := bbl.Down()
		Eventually(session, 10*time.Minute).Should(gexec.Exit())
	})

	It("can tear down a director", func() {
		By("up'ing a new director", func() {
			session := bbl.Up(configuration.IAAS, []string{"--name", bbl.PredefinedEnvID()})
			Eventually(session, 40*time.Minute).Should(gexec.Exit(0))
		})

		By("checking if the bosh director exists", func() {
			directorAddress := bbl.DirectorAddress()
			caCertPath := bbl.SaveDirectorCA()

			exists, err := boshcli.DirectorExists(directorAddress, caCertPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		By("destroying the director", func() {
			session := bbl.Down()
			Eventually(session, 10*time.Minute).Should(gexec.Exit(0))
		})
	})
})
