package gcp_test

import (
	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/application/gcp"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EnvironmentValidator", func() {
	var (
		environmentValidator gcp.EnvironmentValidator
	)

	BeforeEach(func() {
		environmentValidator = gcp.NewEnvironmentValidator()
	})

	Context("when there is a terraform state", func() {
		It("returns no error", func() {
			err := environmentValidator.Validate(storage.State{
				IAAS:    "gcp",
				TFState: "tf-state",
			})

			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when tf state is empty", func() {
		It("returns a BBLNotFound error when tf state is empty", func() {
			err := environmentValidator.Validate(storage.State{
				IAAS:    "gcp",
				TFState: "",
			})

			Expect(err).To(MatchError(application.BBLNotFound))
		})
	})
})
