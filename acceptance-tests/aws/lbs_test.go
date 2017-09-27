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

var _ = Describe("lbs test", func() {
	var (
		bbl actors.BBL
		aws actors.AWS

		certPath       string
		chainPath      string
		keyPath        string
		otherCertPath  string
		otherChainPath string
		otherKeyPath   string
		vpcName        string
	)

	BeforeEach(func() {
		configuration, err := acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "lbs-env")
		aws = actors.NewAWS(configuration)

		certPath, err = testhelpers.WriteContentsToTempFile(testhelpers.BBL_CERT)
		Expect(err).NotTo(HaveOccurred())

		chainPath, err = testhelpers.WriteContentsToTempFile(testhelpers.BBL_CHAIN)
		Expect(err).NotTo(HaveOccurred())

		keyPath, err = testhelpers.WriteContentsToTempFile(testhelpers.BBL_KEY)
		Expect(err).NotTo(HaveOccurred())

		otherCertPath, err = testhelpers.WriteContentsToTempFile(testhelpers.OTHER_BBL_CERT)
		Expect(err).NotTo(HaveOccurred())

		otherKeyPath, err = testhelpers.WriteContentsToTempFile(testhelpers.OTHER_BBL_KEY)
		Expect(err).NotTo(HaveOccurred())

		otherChainPath, err = testhelpers.WriteContentsToTempFile(testhelpers.OTHER_BBL_CHAIN)
		Expect(err).NotTo(HaveOccurred())

		vpcName = fmt.Sprintf("%s-vpc", bbl.PredefinedEnvID())
	})

	AfterEach(func() {
		session := bbl.Destroy()
		Eventually(session, 10*time.Minute).Should(gexec.Exit())
	})

	It("creates, updates and deletes cf LBs with the specified cert and key", func() {
		acceptance.SkipUnless("load-balancers")
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
			session := bbl.LBs()
			Eventually(session, 2*time.Second).Should(gexec.Exit(0))

			stdout := string(session.Out.Contents())
			Expect(stdout).To(MatchRegexp("CF Router LB: .*"))
			Expect(stdout).To(MatchRegexp("CF SSH Proxy LB: .*"))
			Expect(stdout).To(MatchRegexp("CF TCP Router LB: .*"))
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
