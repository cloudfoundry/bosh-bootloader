package integration_test

import (
	"fmt"

	integration "github.com/cloudfoundry/bosh-bootloader/integration-test"
	"github.com/cloudfoundry/bosh-bootloader/integration-test/actors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("no director test", func() {
	var (
		bbl   actors.BBL
		gcp   actors.GCP
		state integration.State

		envID string
	)

	BeforeEach(func() {
		var err error
		configuration, err := integration.LoadGCPConfig()
		Expect(err).NotTo(HaveOccurred())

		state = integration.NewState(configuration.StateFileDir)
		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "no-director-env")
		gcp = actors.NewGCP(configuration)
	})

	It("successfully bbls up and destroys with no director", func() {
		var (
			expectedSSHKey  string
			directorAddress string
		)

		By("calling bbl up", func() {
			bbl.Up(actors.GCPIAAS, []string{"--name", envID, "--no-director"})
		})

		By("checking the ssh key exists", func() {
			expectedSSHKey = fmt.Sprintf("vcap:%s vcap", state.SSHPublicKey())

			actualSSHKeys, err := gcp.SSHKey()
			Expect(err).NotTo(HaveOccurred())
			Expect(actualSSHKeys).To(ContainSubstring(expectedSSHKey))
		})

		By("checking that the bosh director does not exists", func() {
			directorAddress = bbl.DirectorAddress()
			Expect(directorAddress).To(Equal(""))
		})

		By("calling bbl destroy", func() {
			bbl.Destroy()
		})

		By("confirming the ssh key does not exist", func() {
			actualSSHKeys, err := gcp.SSHKey()
			Expect(err).NotTo(HaveOccurred())
			Expect(actualSSHKeys).NotTo(ContainSubstring(expectedSSHKey))
		})

		By("checking the network and subnet do not exist", func() {
			network, _ := gcp.GetNetwork(envID + "-network")
			Expect(network).To(BeNil())

			subnet, _ := gcp.GetSubnet(envID + "-subnet")
			Expect(subnet).To(BeNil())
		})
	})
})
