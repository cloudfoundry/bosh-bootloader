package integration_test

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/pivotal-cf-experimental/bosh-bootloader/integration-test"
	"github.com/pivotal-cf-experimental/bosh-bootloader/integration-test/actors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("load balancer tests", func() {
	var (
		bbl         actors.BBL
		aws         actors.AWS
		state       integration.State
		certBody    []byte
		newCertBody []byte
	)

	BeforeEach(func() {
		var err error
		stateDirectory, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		configuration, err := integration.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(stateDirectory, pathToBBL, configuration)
		aws = actors.NewAWS(configuration)
		state = integration.NewState(stateDirectory)

		certBody, err = ioutil.ReadFile("fixtures/bbl-intermediate.crt")
		Expect(err).NotTo(HaveOccurred())

		newCertBody, err = ioutil.ReadFile("fixtures/new-bbl.crt")
		Expect(err).NotTo(HaveOccurred())

	})

	AfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			bbl.Destroy()
		}
	})

	Context("when bbl up has already created a BOSH director", func() {
		It("creates, updates and deletes an LB with the specified cert and key", func() {
			bbl.Up("")

			stackName := state.StackName()

			Expect(aws.StackExists(stackName)).To(BeTrue())
			Expect(aws.LoadBalancers(stackName)).To(BeEmpty())

			bbl.CreateLB("concourse", "fixtures/bbl-intermediate.crt", "fixtures/bbl-intermediate.key", "fixtures/bbl.crt")

			Expect(aws.LoadBalancers(stackName)).To(HaveKey("ConcourseLoadBalancer"))
			Expect(strings.TrimSpace(aws.DescribeCertificate(state.CertificateName()).Body)).To(Equal(strings.TrimSpace(string(certBody))))

			bbl.UpdateLB("fixtures/new-bbl.crt", "fixtures/new-bbl.key")
			Expect(aws.LoadBalancers(stackName)).To(HaveKey("ConcourseLoadBalancer"))

			certificateName := state.CertificateName()
			Expect(strings.TrimSpace(aws.DescribeCertificate(certificateName).Body)).To(Equal(strings.TrimSpace(string(newCertBody))))

			session := bbl.LBs()
			stdout := session.Out.Contents()
			Expect(stdout).To(ContainSubstring(fmt.Sprintf("Concourse LB: %s", aws.LoadBalancers(stackName)["ConcourseLoadBalancer"])))

			bbl.DeleteLB()
			Expect(aws.LoadBalancers(stackName)).NotTo(HaveKey("ConcourseLoadBalancer"))
			Expect(strings.TrimSpace(aws.DescribeCertificate(certificateName).Body)).To(BeEmpty())

			bbl.Destroy()
		})
	})
})
