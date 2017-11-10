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

var _ = Describe("create lbs test", func() {
	var (
		bbl actors.BBL
		aws actors.AWS

		certPath  string
		chainPath string
		keyPath   string
		vpcName   string
	)

	BeforeEach(func() {
		acceptance.SkipUnless("create-lbs")

		configuration, err := acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "create-lbs-env")
		aws = actors.NewAWS(configuration)

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
		session := bbl.Up("--name", bbl.PredefinedEnvID(), "--no-director")
		Eventually(session, 40*time.Minute).Should(gexec.Exit(0))

		By("verifying there are no load balancers", func() {
			Expect(aws.LoadBalancers(vpcName)).To(BeEmpty())
		})

		By("creating cf lbs", func() {
			session := bbl.CreateLB("cf", certPath, keyPath, chainPath)
			Eventually(session, 10*time.Minute).Should(gexec.Exit(0))

			Expect(aws.LoadBalancers(vpcName)).To(HaveLen(3))
			Expect(aws.LoadBalancers(vpcName)).To(ConsistOf(
				MatchRegexp(".*-cf-router-lb"),
				MatchRegexp(".*-cf-ssh-lb"),
				MatchRegexp(".*-cf-tcp-lb"),
			))
		})

		By("verifying that the bbl lbs output contains the cf lbs", func() {
			stdout := bbl.Lbs()
			Expect(stdout).To(MatchRegexp("CF Router LB: .*"))
			Expect(stdout).To(MatchRegexp("CF SSH Proxy LB: .*"))
			Expect(stdout).To(MatchRegexp("CF TCP Router LB: .*"))
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

		By("confirming that the cf lbs do not exist", func() {
			Expect(aws.LoadBalancers(vpcName)).To(BeEmpty())
		})
	})
})
