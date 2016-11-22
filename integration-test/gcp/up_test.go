package integration_test

import (
	"fmt"

	integration "github.com/cloudfoundry/bosh-bootloader/integration-test"
	"github.com/cloudfoundry/bosh-bootloader/integration-test/actors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("up test", func() {
	var (
		bbl       actors.BBL
		gcp       actors.GCP
		terraform actors.Terraform
		state     integration.State
	)

	BeforeEach(func() {
		var err error
		configuration, err := integration.LoadGCPConfig()
		Expect(err).NotTo(HaveOccurred())

		state = integration.NewState(configuration.StateFileDir)
		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration)
		gcp = actors.NewGCP(configuration)
		terraform = actors.NewTerraform(configuration)
	})

	It("successfully bbls up", func() {
		bbl.Up(actors.GCPIAAS)

		By("creating and uploading a ssh key", func() {
			expectedSSHKey := fmt.Sprintf("vcap:%s vcap\n", state.SSHPublicKey())

			actualSSHKey, err := gcp.SSHKey()
			Expect(err).NotTo(HaveOccurred())
			Expect(actualSSHKey).To(Equal(expectedSSHKey))
		})

		By("creating a network and subnet", func() {
			network, err := gcp.GetNetwork(state.EnvID() + "-network")
			Expect(err).NotTo(HaveOccurred())
			Expect(network).NotTo(BeNil())

			subnet, err := gcp.GetSubnet(state.EnvID() + "-subnet")
			Expect(err).NotTo(HaveOccurred())
			Expect(subnet).NotTo(BeNil())
		})

		By("cleaning up", func() {
			err := terraform.Destroy(state)
			Expect(err).NotTo(HaveOccurred())

			err = gcp.RemoveSSHKey()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
