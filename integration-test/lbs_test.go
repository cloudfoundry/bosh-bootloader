package integration_test

import (
	"fmt"
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
			loadBalancers := aws.LoadBalancers(stackName)

			Expect(aws.StackExists(stackName)).To(BeTrue())
			Expect(loadBalancers).To(BeEmpty())

			bbl.CreateLB("concourse")
			loadBalancers = aws.LoadBalancers(stackName)

			Expect(loadBalancers).To(HaveKey("ConcourseLoadBalancer"))
			Expect(strings.TrimSpace(aws.DescribeCertificate(state.CertificateName()).Body)).To(Equal(strings.TrimSpace(string(certBody))))

			bbl.UpdateLB("bbl-certs/new-bbl.crt", "bbl-certs/new-bbl.key")
			Expect(loadBalancers).To(HaveKey("ConcourseLoadBalancer"))
			Expect(strings.TrimSpace(aws.DescribeCertificate(state.CertificateName()).Body)).To(Equal(strings.TrimSpace(string(newCertBody))))

			session := bbl.LBs()
			stdout := session.Out.Contents()
			Expect(stdout).To(ContainSubstring(fmt.Sprintf("Concourse LB: %s", loadBalancers["ConcourseLoadBalancer"])))

			bbl.Destroy()
		})
	})
})
