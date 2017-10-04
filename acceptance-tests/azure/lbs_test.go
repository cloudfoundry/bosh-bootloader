package acceptance_test

import (
	"time"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = FDescribe("lbs test", func() {
	var (
		bbl actors.BBL
		azure actors.Azure
	)

	BeforeEach(func() {
		configuration, err := acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "lbs-env")
		azure = actors.NewAzure(configuration)

		session := bbl.Up("--name", bbl.PredefinedEnvID(), "--no-director")
		Eventually(session, 40*time.Minute).Should(gexec.Exit(0))
	})

	AfterEach(func() {
		session := bbl.Destroy()
		Eventually(session, 10*time.Minute).Should(gexec.Exit())
	})

	It("successfully creates, updates, and deletes cf lbs", func() {
		By("creating cf load balancers", func() {
			certPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_CERT)
			Expect(err).NotTo(HaveOccurred())

			keyPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_KEY)
			Expect(err).NotTo(HaveOccurred())

			session := bbl.CreateLB("cf", certPath, keyPath, "")
			Eventually(session, 10*time.Minute).Should(gexec.Exit(0))
		})

		By("confirming that the app gateway exist", func() {

			// TODO niroy
			// appGateway, err := azure.GetAppGateway()
			// Expect(err).NotTo(HaveOccurred())
			// Expect(appGateway.something).To(HaveLen(1))
		})

		By("verifying the bbl lbs output", func() {
			session := bbl.LBs()
			Eventually(session, 2*time.Second).Should(gexec.Exit(0))

			stdout := string(session.Out.Contents())
			Expect(stdout).To(MatchRegexp("CF Router LB: .*"))
			Expect(stdout).To(MatchRegexp("CF SSH Proxy LB: .*"))
			Expect(stdout).To(MatchRegexp("CF TCP Router LB: .*"))
			Expect(stdout).To(MatchRegexp("CF WebSocket LB: .*"))
		})

		By("deleting lbs", func() {
			session := bbl.DeleteLBs()
			Eventually(session, 15*time.Minute).Should(gexec.Exit(0))
		})

		By("confirming that the app gateway does not exist", func() {

			// appGateway, err := azure.GetAppGateway()
			// Expect(err).NotTo(HaveOccurred())
			// Expect(appGateway). to be false/not exist
		})
	})
})
