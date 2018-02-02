package compute_test

import (
	"errors"

	"github.com/genevievelesperance/leftovers/gcp/compute"
	"github.com/genevievelesperance/leftovers/gcp/compute/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gcpcompute "google.golang.org/api/compute/v1"
)

var _ = Describe("Addresses", func() {
	var (
		client  *fakes.AddressesClient
		logger  *fakes.Logger
		regions map[string]string

		addresses compute.Addresses
	)

	BeforeEach(func() {
		client = &fakes.AddressesClient{}
		logger = &fakes.Logger{}
		regions = map[string]string{"https://region-1": "region-1"}

		logger.PromptCall.Returns.Proceed = true

		addresses = compute.NewAddresses(client, logger, regions)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			client.ListAddressesCall.Returns.Output = &gcpcompute.AddressList{
				Items: []*gcpcompute.Address{{
					Name:   "banana-address",
					Region: "https://region-1",
				}},
			}
			filter = "banana"
		})

		It("lists, filters, and prompts for addresses to delete", func() {
			list, err := addresses.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListAddressesCall.CallCount).To(Equal(1))
			Expect(client.ListAddressesCall.Receives.Region).To(Equal("region-1"))

			Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to delete address banana-address?"))

			Expect(list).To(HaveLen(1))
			Expect(list).To(HaveKeyWithValue("banana-address", "region-1"))
		})

		Context("when the client fails to list addresses", func() {
			BeforeEach(func() {
				client.ListAddressesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := addresses.List(filter)
				Expect(err).To(MatchError("Listing addresses for region region-1: some error"))
			})
		})

		Context("when the address name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := addresses.List("grape")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.CallCount).To(Equal(0))
				Expect(list).To(HaveLen(0))
			})
		})

		Context("when the address is in use", func() {
			BeforeEach(func() {
				client.ListAddressesCall.Returns.Output = &gcpcompute.AddressList{
					Items: []*gcpcompute.Address{{
						Name:   "banana-address",
						Region: "https://region-1",
						Users:  []string{"a-virtual-machine"},
					}},
				}
			})

			It("does not add it to the list", func() {
				list, err := addresses.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.CallCount).To(Equal(0))
				Expect(list).To(HaveLen(0))
			})
		})

		Context("when the user says no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptCall.Returns.Proceed = false
			})

			It("does not add it to the list", func() {
				list, err := addresses.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})

	Describe("Delete", func() {
		var list map[string]string

		BeforeEach(func() {
			list = map[string]string{"banana-address": "region-1"}
		})

		It("deletes addresses", func() {
			addresses.Delete(list)

			Expect(client.DeleteAddressCall.CallCount).To(Equal(1))
			Expect(client.DeleteAddressCall.Receives.Region).To(Equal("region-1"))
			Expect(client.DeleteAddressCall.Receives.Address).To(Equal("banana-address"))

			Expect(logger.PrintfCall.Messages).To(Equal([]string{"SUCCESS deleting address banana-address\n"}))
		})

		Context("when the client fails to delete the address", func() {
			BeforeEach(func() {
				client.DeleteAddressCall.Returns.Error = errors.New("some error")
			})

			It("logs the error", func() {
				addresses.Delete(list)

				Expect(logger.PrintfCall.Messages).To(Equal([]string{"ERROR deleting address banana-address: some error\n"}))
			})
		})
	})
})
