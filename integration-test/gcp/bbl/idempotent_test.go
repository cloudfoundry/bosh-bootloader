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
	)

	BeforeEach(func() {
		var err error
		configuration, err := integration.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "reentrant-env")
	})

	It("is able to bbl up idempotently with a director", func() {
		bbl.Up(actors.GCPIAAS, []string{"--name", bbl.PredefinedEnvID()})

		bbl.Up(actors.GCPIAAS, []string{})

		bbl.Destroy()
	})

	It("is able to bbl up idempotently with no director", func() {
		bbl.Up(actors.GCPIAAS, []string{"--name", bbl.PredefinedEnvID(), "--no-director"})

		bbl.Up(actors.GCPIAAS, []string{})

		bbl.Destroy()
	})
})
