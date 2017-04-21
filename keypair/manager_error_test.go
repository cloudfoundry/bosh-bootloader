package keypair_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/keypair"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ManagerError", func() {
	Describe("Error", func() {
		It("returns the internal error message", func() {
			managerError := keypair.NewManagerError(storage.State{}, errors.New("some internal error"))
			Expect(managerError.Error()).To(Equal("some internal error"))
		})
	})

	Describe("BBLState", func() {
		It("returns the bbl state provided", func() {
			managerError := keypair.NewManagerError(storage.State{
				IAAS: "gcp",
			}, errors.New("some internal error"))

			actualBBLState := managerError.BBLState()
			Expect(actualBBLState).To(Equal(storage.State{
				IAAS: "gcp",
			}))
		})
	})
})
