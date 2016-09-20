package bosh_test

import (
	"github.com/cloudfoundry/bosh-bootloader/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CompilationGenerator", func() {
	Describe("Generate", func() {
		It("returns a slice of disk types for cloud config", func() {
			generator := bosh.NewCompilationGenerator()
			compilation := generator.Generate()

			Expect(compilation).To(Equal(
				&bosh.Compilation{
					Workers:             6,
					Network:             "private",
					AZ:                  "z1",
					ReuseCompilationVMs: true,
					VMType:              "c3.large",
					VMExtensions: []string{
						"100GB_ephemeral_disk",
					},
				},
			))
		})
	})
})
