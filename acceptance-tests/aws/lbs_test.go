package acceptance_test

import (
	"fmt"
	"strings"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("lbs test", func() {
	var (
		bbl     actors.BBL
		aws     actors.AWS
		bosh    actors.BOSH
		boshcli actors.BOSHCLI
		state   acceptance.State

		certPath       string
		chainPath      string
		keyPath        string
		otherCertPath  string
		otherChainPath string
		otherKeyPath   string
		vpcName        string
	)

	BeforeEach(func() {
		var err error
		configuration, err := acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "lbs-env")
		aws = actors.NewAWS(configuration)
		bosh = actors.NewBOSH()
		boshcli = actors.NewBOSHCLI()
		state = acceptance.NewState(configuration.StateFileDir)

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

		bbl.Up(actors.AWSIAAS, []string{"--name", bbl.PredefinedEnvID(), "--no-director"})

		vpcName = fmt.Sprintf("%s-vpc", bbl.PredefinedEnvID())
	})

	AfterEach(func() {
		if !CurrentGinkgoTestDescription().Failed {
			bbl.Destroy()
		}
	})

	It("creates, updates and deletes a concourse LB with the specified cert and key", func() {
		By("verifying there are no load balancers", func() {
			Expect(aws.LoadBalancers(vpcName)).To(BeEmpty())
		})

		By("creating a concourse lb", func() {
			bbl.CreateLB("concourse", certPath, keyPath, chainPath)

			Expect(aws.LoadBalancers(vpcName)).To(HaveLen(1))
			Expect(aws.LoadBalancers(vpcName)).To(Equal([]string{fmt.Sprintf("%s-concourse-lb", bbl.PredefinedEnvID())}))

			certificateName := aws.GetSSLCertificateNameFromLBs(bbl.PredefinedEnvID())
			Expect(strings.TrimSpace(aws.DescribeCertificate(certificateName).Body)).To(Equal(strings.TrimSpace(testhelpers.BBL_CERT)))
		})

		By("verifying that the bbl lbs output contains the concourse lb", func() {
			session := bbl.LBs()
			stdout := string(session.Out.Contents())
			Expect(stdout).To(MatchRegexp("Concourse LB: .*"))
		})

		By("updating the certs of the lb", func() {
			bbl.UpdateLB(otherCertPath, otherKeyPath, otherChainPath)
			Expect(aws.LoadBalancers(vpcName)).To(HaveLen(1))
			Expect(aws.LoadBalancers(vpcName)).To(Equal([]string{fmt.Sprintf("%s-concourse-lb", bbl.PredefinedEnvID())}))

			certificateName := aws.GetSSLCertificateNameFromLBs(bbl.PredefinedEnvID())
			Expect(strings.TrimSpace(aws.DescribeCertificate(certificateName).Body)).To(Equal(strings.TrimSpace(string(testhelpers.OTHER_BBL_CERT))))
		})

		By("deleting lbs", func() {
			bbl.DeleteLBs()
		})

		By("confirming that the concourse lb does not exist", func() {
			Expect(aws.LoadBalancers(vpcName)).To(BeEmpty())
		})
	})

	It("creates, updates and deletes cf LBs with the specified cert and key", func() {
		By("verifying there are no load balancers", func() {
			Expect(aws.LoadBalancers(vpcName)).To(BeEmpty())
		})

		By("creating cf lbs", func() {
			bbl.CreateLB("cf", certPath, keyPath, chainPath)

			Expect(aws.LoadBalancers(vpcName)).To(HaveLen(3))
			Expect(aws.LoadBalancers(vpcName)).To(ConsistOf(
				MatchRegexp(".*-cf-router-lb"),
				MatchRegexp(".*-cf-ssh-lb"),
				MatchRegexp(".*-cf-tcp-lb"),
			))

			certificateName := aws.GetSSLCertificateNameFromLBs(bbl.PredefinedEnvID())
			Expect(strings.TrimSpace(aws.DescribeCertificate(certificateName).Body)).To(Equal(strings.TrimSpace(testhelpers.BBL_CERT)))
		})

		By("verifying that the bbl lbs output contains the cf lbs", func() {
			session := bbl.LBs()
			stdout := string(session.Out.Contents())
			Expect(stdout).To(MatchRegexp("CF Router LB: .*"))
			Expect(stdout).To(MatchRegexp("CF SSH Proxy LB: .*"))
			Expect(stdout).To(MatchRegexp("CF TCP Router LB: .*"))
		})

		By("updating the certs of the cf router lb", func() {
			bbl.UpdateLB(otherCertPath, otherKeyPath, otherChainPath)
			Expect(aws.LoadBalancers(vpcName)).To(HaveLen(3))
			Expect(aws.LoadBalancers(vpcName)).To(ConsistOf(
				MatchRegexp(".*-cf-router-lb"),
				MatchRegexp(".*-cf-ssh-lb"),
				MatchRegexp(".*-cf-tcp-lb"),
			))

			certificateName := aws.GetSSLCertificateNameFromLBs(bbl.PredefinedEnvID())
			Expect(strings.TrimSpace(aws.DescribeCertificate(certificateName).Body)).To(Equal(strings.TrimSpace(string(testhelpers.OTHER_BBL_CERT))))
		})

		By("verifying that the bbl lbs output contains the cf lbs", func() {
			session := bbl.LBs()
			stdout := string(session.Out.Contents())
			Expect(stdout).To(MatchRegexp("CF Router LB: .*"))
			Expect(stdout).To(MatchRegexp("CF SSH Proxy LB: .*"))
			Expect(stdout).To(MatchRegexp("CF TCP Router LB: .*"))
		})
	})
})
