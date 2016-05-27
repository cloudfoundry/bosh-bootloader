package integration_test

import (
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/integration-test"
	"github.com/pivotal-cf-experimental/bosh-bootloader/integration-test/actors"
)

var _ = Describe("up and destroy", func() {
	var bbl actors.BBL
	var aws actors.AWS
	var state integration.State
	var bosh actors.BOSH

	BeforeEach(func() {
		stateDirectory, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		configuration, err := integration.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(stateDirectory, pathToBBL, configuration)
		aws = actors.NewAWS(configuration)
		state = integration.NewState(stateDirectory)
		bosh = actors.NewBOSH()
	})

	AfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			bbl.Destroy()
		}
	})

	It("creates and destroys a stack with a bosh director and load balancers", func() {
		bbl.Up("")
		originalStateChecksum := state.Checksum()
		stackName := state.StackName()
		directorAddress := bbl.DirectorAddress()
		directorUsername := bbl.DirectorUsername()
		directorPassword := bbl.DirectorPassword()

		Expect(aws.StackExists(stackName)).To(BeTrue())
		Expect(aws.LoadBalancers(stackName)).To(BeEmpty())
		Expect(bosh.DirectorExists(directorAddress, directorUsername, directorPassword)).To(BeTrue())

		bbl.Destroy()
		Expect(bosh.DirectorExists(directorAddress, directorUsername, directorPassword)).To(BeFalse())
		Expect(aws.StackExists(stackName)).To(BeFalse())
	})
})
