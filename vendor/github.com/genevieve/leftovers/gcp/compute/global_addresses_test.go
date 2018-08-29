package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gcpcompute "google.golang.org/api/compute/v1"
)

var _ = Describe("GlobalAddresses", func() {
	var (
		client *fakes.GlobalAddressesClient
		logger *fakes.Logger

		addresses compute.GlobalAddresses
	)

	BeforeEach(func() {
		client = &fakes.GlobalAddressesClient{}
		logger = &fakes.Logger{}

		logger.PromptWithDetailsCall.Returns.Proceed = true

		addresses = compute.NewGlobalAddresses(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			client.ListGlobalAddressesCall.Returns.Output = []*gcpcompute.Address{{
				Name: "banana-address",
			}}
			filter = "banana"
		})

		It("lists, filters, and prompts for addresses to delete", func() {
			list, err := addresses.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListGlobalAddressesCall.CallCount).To(Equal(1))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("Global Address"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-address"))

			Expect(list).To(HaveLen(1))
		})

		Context("when the client fails to list addresses", func() {
			BeforeEach(func() {
				client.ListGlobalAddressesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := addresses.List(filter)
				Expect(err).To(MatchError("List Global Addresses: some error"))
			})
		})

		Context("when the address name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := addresses.List("grape")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(0))
				Expect(list).To(HaveLen(0))
			})
		})

		Context("when the user says no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("does not add it to the list", func() {
				list, err := addresses.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})
})
