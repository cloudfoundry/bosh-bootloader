package acceptance_test

import (
	"fmt"
	"time"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("plan lbs test", func() {
	var (
		bbl actors.BBL
		gcp actors.GCP

		certPath  string
		chainPath string
		keyPath   string
		vpcName   string
	)

	BeforeEach(func() {
		configuration, err := acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "plan-lbs-env")
		gcp = actors.NewGCP(configuration)

		certPath, err = testhelpers.WriteContentsToTempFile(testhelpers.BBL_CERT)
		Expect(err).NotTo(HaveOccurred())
		chainPath, err = testhelpers.WriteContentsToTempFile(testhelpers.BBL_CHAIN)
		Expect(err).NotTo(HaveOccurred())
		keyPath, err = testhelpers.WriteContentsToTempFile(testhelpers.BBL_KEY)
		Expect(err).NotTo(HaveOccurred())

		vpcName = fmt.Sprintf("%s-vpc", bbl.PredefinedEnvID())
	})

	AfterEach(func() {
		session := bbl.Destroy()
		Eventually(session, 10*time.Minute).Should(gexec.Exit())
	})

	It("creates, updates and deletes cf LBs with the specified cert and key", func() {
		session := bbl.Up(
			"--name", bbl.PredefinedEnvID(),
			"--no-director",
			"--lb-type", "cf",
			"--lb-cert", certPath,
			"--lb-key", keyPath,
		)
		Eventually(session, 40*time.Minute).Should(gexec.Exit(0))

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

		By("verifying that vm extensions were added to the cloud config", func() {
			cloudConfig := bbl.CloudConfig()
			vmExtensions := acceptance.VmExtensionNames(cloudConfig)
			Expect(vmExtensions).To(ContainElement("cf-router-network-properties"))
			Expect(vmExtensions).To(ContainElement("diego-ssh-proxy-network-properties"))
			Expect(vmExtensions).To(ContainElement("cf-tcp-router-network-properties"))
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

		By("confirming that the target pools do not exist", func() {
			targetPools := []string{bbl.PredefinedEnvID() + "-cf-ssh-proxy", bbl.PredefinedEnvID() + "-cf-tcp-router"}
			for _, p := range targetPools {
				_, err := gcp.GetTargetPool(p)
				Expect(err).To(MatchError(MatchRegexp(`The resource 'projects\/.+` + p + `' was not found`)))
			}
		})
	})
})
