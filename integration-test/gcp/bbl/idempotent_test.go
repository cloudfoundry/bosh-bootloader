package integration_test

import (
	integration "github.com/cloudfoundry/bosh-bootloader/integration-test"
	"github.com/cloudfoundry/bosh-bootloader/integration-test/actors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("idempotent test", func() {
	var (
		bbl actors.BBL

		envID string
	)

	BeforeEach(func() {
		var err error
		configuration, err := integration.LoadGCPConfig()
		Expect(err).NotTo(HaveOccurred())

		envID = configuration.GCPEnvPrefix + "bbl-ci-reentrant-env"
		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, envID)
	})

	It("is able to bbl up idempotently with a director", func() {
		bbl.Up(actors.GCPIAAS, []string{"--name", envID})

		bbl.Up(actors.GCPIAAS, []string{})

		bbl.Destroy()
	})

	It("is able to bbl up idempotently with no director", func() {
		bbl.Up(actors.GCPIAAS, []string{"--name", envID, "--no-director"})

		bbl.Up(actors.GCPIAAS, []string{})

		bbl.Destroy()
	})
})
