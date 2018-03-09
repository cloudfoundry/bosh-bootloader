package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Address", func() {
	var (
		client *fakes.AddressesClient
		name   string
		region string

		address compute.Address
	)

	BeforeEach(func() {
		client = &fakes.AddressesClient{}
		name = "banana"
		region = "us-banana"

		address = compute.NewAddress(client, name, region)
	})

	Describe("Delete", func() {
		It("deletes the address", func() {
			err := address.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteAddressCall.CallCount).To(Equal(1))
			Expect(client.DeleteAddressCall.Receives.Address).To(Equal(name))
			Expect(client.DeleteAddressCall.Receives.Region).To(Equal(region))
		})

		Context("when the client fails to delete", func() {
			BeforeEach(func() {
				client.DeleteAddressCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := address.Delete()
				Expect(err).To(MatchError("ERROR deleting address banana: the-error"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(address.Name()).To(Equal(name))
		})
	})

	Describe("Type", func() {
		It("returns \"address\"", func() {
			Expect(address.Type()).To(Equal("address"))
		})
	})
})
