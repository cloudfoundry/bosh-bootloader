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
		users := 0

		address = compute.NewAddress(client, name, region, users)
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
				Expect(err).To(MatchError("Delete: the-error"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(address.Name()).To(Equal(name))
		})

		Context("when the address is in use", func() {
			BeforeEach(func() {
				users := 1
				address = compute.NewAddress(client, name, region, users)
			})
			It("adds the number of users to the name", func() {
				Expect(address.Name()).To(Equal("banana (Users:1)"))
			})
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(address.Type()).To(Equal("Address"))
		})
	})
})
