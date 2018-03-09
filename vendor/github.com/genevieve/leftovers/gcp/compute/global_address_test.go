package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GlobalAddress", func() {
	var (
		client *fakes.GlobalAddressesClient
		name   string

		globalAddress compute.GlobalAddress
	)

	BeforeEach(func() {
		client = &fakes.GlobalAddressesClient{}
		name = "banana"

		globalAddress = compute.NewGlobalAddress(client, name)
	})

	Describe("Delete", func() {
		It("deletes the global address", func() {
			err := globalAddress.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteGlobalAddressCall.CallCount).To(Equal(1))
			Expect(client.DeleteGlobalAddressCall.Receives.Address).To(Equal(name))
		})

		Context("when the client fails to delete", func() {
			BeforeEach(func() {
				client.DeleteGlobalAddressCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := globalAddress.Delete()
				Expect(err).To(MatchError("ERROR deleting global address banana: the-error"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(globalAddress.Name()).To(Equal(name))
		})
	})

	Describe("Type", func() {
		It("returns \"global address\"", func() {
			Expect(globalAddress.Type()).To(Equal("global address"))
		})
	})
})
