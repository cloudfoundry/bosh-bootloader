package acceptance_test

import (
	"encoding/base64"
	"time"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("lbs test", func() {
	var (
		bbl   actors.BBL
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
			pfx_data, err := base64.StdEncoding.DecodeString(testhelpers.PFX_BASE64)
			Expect(err).NotTo(HaveOccurred())

			certPath, err := testhelpers.WriteByteContentsToTempFile(pfx_data)
			Expect(err).NotTo(HaveOccurred())

			keyPath, err := testhelpers.WriteContentsToTempFile(testhelpers.PFX_PASSWORD)
			Expect(err).NotTo(HaveOccurred())

			session := bbl.CreateLB("cf", certPath, keyPath, "")
			Eventually(session, 10*time.Minute).Should(gexec.Exit(0))
		})

		By("confirming that the app gateway exist", func() {
			exists, err := azure.GetApplicationGateway(bbl.PredefinedEnvID(), "some-gateway")
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		By("verifying the bbl lbs output", func() {
			stdout := bbl.Lbs()
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
			exists, err := azure.GetApplicationGateway(bbl.PredefinedEnvID(), "some-gateway")
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeFalse())
		})
	})
})
