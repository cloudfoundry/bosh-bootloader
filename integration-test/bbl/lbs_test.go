package integration_test

import (
	"fmt"
	"strings"

	"github.com/pivotal-cf-experimental/bosh-bootloader/integration-test"
	"github.com/pivotal-cf-experimental/bosh-bootloader/integration-test/actors"
	"github.com/pivotal-cf-experimental/bosh-bootloader/testhelpers"

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
	)

	BeforeEach(func() {
		var err error
		configuration, err := integration.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration)
		aws = actors.NewAWS(configuration)
		bosh = actors.NewBOSH()
		boshcli = actors.NewBOSHCLI()
		state = integration.NewState(configuration.StateFileDir)

	})

	It("creates, updates and deletes an LB with the specified cert and key", func() {
		bbl.Up()

		stackName := state.StackName()
		directorAddress := bbl.DirectorAddress()
		caCertPath := bbl.SaveDirectorCA()

		Expect(aws.StackExists(stackName)).To(BeTrue())
		Expect(aws.LoadBalancers(stackName)).To(BeEmpty())
		Expect(boshcli.DirectorExists(directorAddress, caCertPath)).To(BeTrue())

		certPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_CERT)
		Expect(err).NotTo(HaveOccurred())

		chainPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_CHAIN)
		Expect(err).NotTo(HaveOccurred())

		keyPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_KEY)
		Expect(err).NotTo(HaveOccurred())

		otherCertPath, err := testhelpers.WriteContentsToTempFile(testhelpers.OTHER_BBL_CERT)
		Expect(err).NotTo(HaveOccurred())

		otherKeyPath, err := testhelpers.WriteContentsToTempFile(testhelpers.OTHER_BBL_KEY)
		Expect(err).NotTo(HaveOccurred())

		bbl.CreateLB("concourse", certPath, keyPath, chainPath)

		Expect(aws.LoadBalancers(stackName)).To(HaveKey("ConcourseLoadBalancer"))
		Expect(strings.TrimSpace(aws.DescribeCertificate(state.CertificateName()).Body)).To(Equal(strings.TrimSpace(testhelpers.BBL_CERT)))

		bbl.UpdateLB(otherCertPath, otherKeyPath)
		Expect(aws.LoadBalancers(stackName)).To(HaveKey("ConcourseLoadBalancer"))

		certificateName := state.CertificateName()
		Expect(strings.TrimSpace(aws.DescribeCertificate(certificateName).Body)).To(Equal(strings.TrimSpace(string(testhelpers.OTHER_BBL_CERT))))

		session := bbl.LBs()
		stdout := session.Out.Contents()
		Expect(stdout).To(ContainSubstring(fmt.Sprintf("Concourse LB: %s", aws.LoadBalancers(stackName)["ConcourseLoadBalancer"])))

		bbl.DeleteLB()
		Expect(aws.LoadBalancers(stackName)).NotTo(HaveKey("ConcourseLoadBalancer"))
		Expect(strings.TrimSpace(aws.DescribeCertificate(certificateName).Body)).To(BeEmpty())

		bbl.Destroy()
		Expect(boshcli.DirectorExists(directorAddress, caCertPath)).To(BeFalse())
		Expect(aws.StackExists(stackName)).To(BeFalse())
	})
})
