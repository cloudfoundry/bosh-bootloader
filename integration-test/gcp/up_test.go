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
		bbl   actors.BBL
		gcp   actors.GCP
		state integration.State
	)

	BeforeEach(func() {
		var err error
		configuration, err := integration.LoadGCPConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration)
		gcp = actors.NewGCP(configuration)
		state = integration.NewState(configuration.StateFileDir)
	})

	It("creates and uploads a ssh key", func() {
		bbl.Up(actors.GCPIAAS)

		expectedSSHKey := fmt.Sprintf("vcap:%s vcap\n", state.SSHPublicKey())

		actualSSHKey, err := gcp.SSHKey()
		Expect(err).NotTo(HaveOccurred())
		Expect(actualSSHKey).To(Equal(expectedSSHKey))

		err = gcp.RemoveSSHKey()
		Expect(err).NotTo(HaveOccurred())
	})
})
