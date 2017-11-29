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
		aws actors.AWS

		certPath  string
		chainPath string
		keyPath   string
		vpcName   string
	)

	BeforeEach(func() {
		acceptance.SkipUnless("plan-lbs")

		configuration, err := acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "plan-lbs-env")
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
		session := bbl.Up(
			"--name", bbl.PredefinedEnvID(),
			"--no-director",
			"--lb-type", "cf",
			"--lb-cert", certPath,
			"--lb-key", keyPath,
			"--lb-chain", chainPath,
		)
		Eventually(session, 40*time.Minute).Should(gexec.Exit(0))

		By("verifying that there are load balancers", func() {
			Expect(aws.LoadBalancers(vpcName)).To(HaveLen(3))
			Expect(aws.LoadBalancers(vpcName)).To(ConsistOf(
				MatchRegexp(".*-cf-router-lb"),
				MatchRegexp(".*-cf-ssh-lb"),
				MatchRegexp(".*-cf-tcp-lb"),
			))
		})

		By("verifying that vm extensions were added to the cloud config", func() {
			cloudConfig := bbl.CloudConfig()
			vmExtensions := acceptance.VmExtensionNames(cloudConfig)
			Expect(vmExtensions).To(ContainElement("cf-router-network-properties"))
			Expect(vmExtensions).To(ContainElement("diego-ssh-proxy-network-properties"))
			Expect(vmExtensions).To(ContainElement("cf-tcp-router-network-properties"))
		})

		By("verifying that the bbl lbs output contains the cf lbs", func() {
			stdout := bbl.Lbs()
			Expect(stdout).To(MatchRegexp("CF SSH Proxy LB: .*"))
			Expect(stdout).To(MatchRegexp("CF TCP Router LB: .*"))
			Expect(stdout).To(MatchRegexp("CF Router LB: .*"))
		})

		By("deleting lbs", func() {
			session := bbl.DeleteLBs()
			Eventually(session, 15*time.Minute).Should(gexec.Exit(0))
		})

		By("confirming that the cf lbs do not exist", func() {
			Expect(aws.LoadBalancers(vpcName)).To(BeEmpty())
		})
	})

	It("creates, updates and deletes concourse LB with the specified cert and key", func() {
		session := bbl.Up(
			"--name", bbl.PredefinedEnvID(),
			"--no-director",
			"--lb-type", "concourse",
			"--lb-cert", certPath,
			"--lb-key", keyPath,
		)
		Eventually(session, 40*time.Minute).Should(gexec.Exit(0))

		By("verifying that there are load balancers", func() {
			Expect(aws.NetworkLoadBalancers(vpcName)).To(HaveLen(1))
			Expect(aws.NetworkLoadBalancers(vpcName)).To(ConsistOf(
				MatchRegexp(".*-concourse-lb"),
			))
		})

		By("verifying that vm extensions were added to the cloud config", func() {
			cloudConfig := bbl.CloudConfig()
			vmExtensions := acceptance.VmExtensionNames(cloudConfig)
			Expect(vmExtensions).To(ContainElement("lb"))
		})

		By("verifying that the bbl lbs output contains the concourse lb", func() {
			stdout := bbl.Lbs()
			Expect(stdout).To(MatchRegexp("Concourse LB: .*"))
		})

		By("deleting lbs", func() {
			session := bbl.DeleteLBs()
			Eventually(session, 15*time.Minute).Should(gexec.Exit(0))
		})

		By("confirming that the concourse lb does not exist", func() {
			Expect(aws.NetworkLoadBalancers(vpcName)).To(BeEmpty())
		})
	})
})
