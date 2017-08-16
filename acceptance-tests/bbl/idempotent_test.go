package acceptance_test

import (
	"time"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("idempotent test", func() {
	var (
		bbl           actors.BBL
		configuration acceptance.Config
	)

	BeforeEach(func() {
		var err error
		configuration, err = acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "reentrant-env")
	})

	AfterEach(func() {
		session := bbl.Destroy()
		Eventually(session, 10*time.Minute).Should(gexec.Exit())
	})

	It("is able to bbl up idempotently with a director", func() {
		session := bbl.Up(configuration.IAAS, []string{"--name", bbl.PredefinedEnvID()})
		Eventually(session, 40*time.Minute).Should(gexec.Exit(0))

		session = bbl.Up(configuration.IAAS, []string{})
		Eventually(session, 40*time.Minute).Should(gexec.Exit(0))
	})

	It("is able to bbl up idempotently with no director", func() {
		session := bbl.Up(configuration.IAAS, []string{"--name", bbl.PredefinedEnvID(), "--no-director"})
		Eventually(session, 40*time.Minute).Should(gexec.Exit(0))

		session = bbl.Up(configuration.IAAS, []string{})
		Eventually(session, 40*time.Minute).Should(gexec.Exit(0))
	})
})
