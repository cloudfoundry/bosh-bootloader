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

var _ = Describe("create lbs test", func() {
	var (
		bbl actors.BBL
		gcp actors.GCP
	)

	BeforeEach(func() {
		configuration, err := acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "create-lbs-env")
		gcp = actors.NewGCP(configuration)

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

		By("confirming that target pools exist", func() {
			targetPools := []string{bbl.PredefinedEnvID() + "-cf-ssh-proxy", bbl.PredefinedEnvID() + "-cf-tcp-router"}
			for _, p := range targetPools {
				targetPool, err := gcp.GetTargetPool(p)
				Expect(err).NotTo(HaveOccurred())
				Expect(targetPool.Name).NotTo(BeNil())
				Expect(targetPool.Name).To(Equal(p))
			}

			targetHTTPSProxy, err := gcp.GetTargetHTTPSProxy(bbl.PredefinedEnvID() + "-https-proxy")
			Expect(err).NotTo(HaveOccurred())
			Expect(targetHTTPSProxy.SslCertificates).To(HaveLen(1))
		})

		By("verifying the bbl lbs output", func() {
			stdout := bbl.Lbs()
			Expect(stdout).To(MatchRegexp("CF Router LB: .*"))
			Expect(stdout).To(MatchRegexp("CF SSH Proxy LB: .*"))
			Expect(stdout).To(MatchRegexp("CF TCP Router LB: .*"))
			Expect(stdout).To(MatchRegexp("CF WebSocket LB: .*"))
		})

		By("verifying that vm extensions were added to the cloud config", func() {
			cloudConfig := bbl.CloudConfig()
			vmExtensions := acceptance.VmExtensionNames(cloudConfig)
			Expect(vmExtensions).To(ContainElement("cf-router-network-properties"))
			Expect(vmExtensions).To(ContainElement("diego-ssh-proxy-network-properties"))
			Expect(vmExtensions).To(ContainElement("cf-tcp-router-network-properties"))
		})

		By("deleting lbs", func() {
			session := bbl.DeleteLBs()
			Eventually(session, 15*time.Minute).Should(gexec.Exit(0))
		})

		By("confirming that the target pools do not exist", func() {
			targetPools := []string{bbl.PredefinedEnvID() + "-cf-ssh-proxy", bbl.PredefinedEnvID() + "-cf-tcp-router"}
			for _, p := range targetPools {
				_, err := gcp.GetTargetPool(p)
				Expect(err).To(MatchError(MatchRegexp(`The resource 'projects\/.+` + p + `' was not found`)))
			}
		})
	})
})
