package integration_test

import (
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/integration-test"
	"github.com/pivotal-cf-experimental/bosh-bootloader/integration-test/actors"
)

var _ = Describe("create-lb", func() {
	var bbl actors.BBL
	var aws actors.AWS
	var state integration.State

	BeforeEach(func() {
		stateDirectory, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		configuration, err := integration.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(stateDirectory, pathToBBL, configuration)
		aws = actors.NewAWS(configuration)
		state = integration.NewState(stateDirectory)
	})

	AfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			bbl.Destroy()
		}
	})

	Context("when bbl up has already created a BOSH director", func() {
		It("creates an LB with specified cert and keys", func() {
			bbl.Up("")

			stackName := state.StackName()
			Expect(aws.StackExists(stackName)).To(BeTrue())
			Expect(aws.LoadBalancers(stackName)).To(BeEmpty())

			bbl.CreateLB("concourse")

			Expect(aws.LoadBalancers(stackName)).To(Equal([]string{"ConcourseLoadBalancer"}))
		})
	})

})
