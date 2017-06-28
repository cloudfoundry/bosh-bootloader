package integration_test

import (
	"fmt"
	"strings"

	integration "github.com/cloudfoundry/bosh-bootloader/integration-test"
	"github.com/cloudfoundry/bosh-bootloader/integration-test/actors"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("load balancer tests", func() {
	var (
		bbl     actors.BBL
		aws     actors.AWS
		bosh    actors.BOSH
		boshcli actors.BOSHCLI
		state   integration.State

		certPath       string
		chainPath      string
		keyPath        string
		otherCertPath  string
		otherChainPath string
		otherKeyPath   string
	)

	BeforeEach(func() {
		var err error
		configuration, err := integration.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "lbs-env")
		aws = actors.NewAWS(configuration)
		bosh = actors.NewBOSH()
		boshcli = actors.NewBOSHCLI()
		state = integration.NewState(configuration.StateFileDir)

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

		bbl.Up(actors.AWSIAAS, []string{"--name", bbl.PredefinedEnvID()})
	})

	AfterEach(func() {
		if !CurrentGinkgoTestDescription().Failed {
			bbl.DeleteLBs()
			bbl.Destroy()
		}
	})

	It("creates, updates and deletes a concourse LB with the specified cert and key", func() {
		lbName := fmt.Sprintf("%s-concourse-lb", bbl.PredefinedEnvID())

		By("verifying there are no load balancers", func() {
			Expect(aws.LoadBalancers()).To(BeEmpty())
		})

		By("creating a concourse lb", func() {
			bbl.CreateLB("concourse", certPath, keyPath, chainPath)

			Expect(aws.LoadBalancers()).To(HaveLen(1))
			Expect(aws.LoadBalancers()).To(Equal([]string{fmt.Sprintf("%s-concourse-lb", bbl.PredefinedEnvID())}))

			certificateName := aws.GetSSLCertificateNameByLoadBalancer(lbName)
			Expect(strings.TrimSpace(aws.DescribeCertificate(certificateName).Body)).To(Equal(strings.TrimSpace(testhelpers.BBL_CERT)))
		})

		By("updating the certs of the lb", func() {
			bbl.UpdateLB(otherCertPath, otherKeyPath, otherChainPath)
			Expect(aws.LoadBalancers()).To(HaveLen(1))
			Expect(aws.LoadBalancers()).To(Equal([]string{fmt.Sprintf("%s-concourse-lb", bbl.PredefinedEnvID())}))

			certificateName := aws.GetSSLCertificateNameByLoadBalancer(lbName)
			Expect(strings.TrimSpace(aws.DescribeCertificate(certificateName).Body)).To(Equal(strings.TrimSpace(string(testhelpers.OTHER_BBL_CERT))))
		})

		By("verifying that the bbl lbs output contains the concourse lb", func() {
			session := bbl.LBs()
			stdout := string(session.Out.Contents())
			Expect(stdout).To(MatchRegexp("Concourse LB: .*"))
		})
	})

	It("creates, updates and deletes cf LBs with the specified cert and key", func() {
		routerLBName := fmt.Sprintf("%s-cf-router-lb", bbl.PredefinedEnvID())

		By("verifying there are no load balancers", func() {
			Expect(aws.LoadBalancers()).To(BeEmpty())
		})

		By("creating cf lbs", func() {
			bbl.CreateLB("cf", certPath, keyPath, chainPath)

			Expect(aws.LoadBalancers()).To(HaveLen(3))
			Expect(aws.LoadBalancers()).To(ConsistOf([]string{
				fmt.Sprintf("%s-cf-router-lb", bbl.PredefinedEnvID()),
				fmt.Sprintf("%s-cf-ssh-lb", bbl.PredefinedEnvID()),
				fmt.Sprintf("%s-cf-tcp-lb", bbl.PredefinedEnvID()),
			}))

			certificateName := aws.GetSSLCertificateNameByLoadBalancer(routerLBName)
			Expect(strings.TrimSpace(aws.DescribeCertificate(certificateName).Body)).To(Equal(strings.TrimSpace(testhelpers.BBL_CERT)))
		})

		By("updating the certs of the cf router lb", func() {
			bbl.UpdateLB(otherCertPath, otherKeyPath, otherChainPath)
			Expect(aws.LoadBalancers()).To(HaveLen(3))
			Expect(aws.LoadBalancers()).To(ConsistOf([]string{
				fmt.Sprintf("%s-cf-router-lb", bbl.PredefinedEnvID()),
				fmt.Sprintf("%s-cf-ssh-lb", bbl.PredefinedEnvID()),
				fmt.Sprintf("%s-cf-tcp-lb", bbl.PredefinedEnvID()),
			}))

			certificateName := aws.GetSSLCertificateNameByLoadBalancer(routerLBName)
			Expect(strings.TrimSpace(aws.DescribeCertificate(certificateName).Body)).To(Equal(strings.TrimSpace(string(testhelpers.OTHER_BBL_CERT))))
		})

		By("verifying that the bbl lbs output contains the cf lbs", func() {
			session := bbl.LBs()
			stdout := string(session.Out.Contents())
			Expect(stdout).To(MatchRegexp("CF Router LB: .*"))
			Expect(stdout).To(MatchRegexp("CF SSH Proxy LB: .*"))
			//Expect(stdout).To(MatchRegexp("CF TCP Router LB: .*"))
		})
	})
})
