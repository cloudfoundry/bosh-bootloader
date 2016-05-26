package integration_test

import (
	"io/ioutil"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/integration-test"
	"github.com/pivotal-cf-experimental/bosh-bootloader/integration-test/actors"
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

		certBody, err = ioutil.ReadFile("bbl-certs/bbl-intermediate.crt")
		Expect(err).NotTo(HaveOccurred())

		newCertBody, err = ioutil.ReadFile("bbl-certs/new-bbl.crt")
		Expect(err).NotTo(HaveOccurred())

	})

	AfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			bbl.Destroy()
		}
	})

	Context("when bbl up has already created a BOSH director", func() {
		It("creates and updates an LB with the specified cert and key", func() {
			bbl.Up("")

			stackName := state.StackName()
			Expect(aws.StackExists(stackName)).To(BeTrue())
			Expect(aws.LoadBalancers(stackName)).To(BeEmpty())

			bbl.CreateLB("concourse")

			Expect(aws.LoadBalancers(stackName)).To(Equal([]string{"ConcourseLoadBalancer"}))
			Expect(strings.TrimSpace(aws.DescribeCertificate(state.CertificateName()).Body)).To(Equal(strings.TrimSpace(string(certBody))))

			bbl.UpdateLB("bbl-certs/new-bbl.crt", "bbl-certs/new-bbl.key")
			Expect(aws.LoadBalancers(stackName)).To(Equal([]string{"ConcourseLoadBalancer"}))
			Expect(strings.TrimSpace(aws.DescribeCertificate(state.CertificateName()).Body)).To(Equal(strings.TrimSpace(string(newCertBody))))

			bbl.Destroy()
		})
	})
})
