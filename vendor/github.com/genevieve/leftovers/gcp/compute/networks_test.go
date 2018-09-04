package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gcpcompute "google.golang.org/api/compute/v1"
)

var _ = Describe("Networks", func() {
	var (
		client *fakes.NetworksClient
		logger *fakes.Logger

		networks compute.Networks
	)

	BeforeEach(func() {
		client = &fakes.NetworksClient{}
		logger = &fakes.Logger{}

		networks = compute.NewNetworks(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.ListNetworksCall.Returns.Output = []*gcpcompute.Network{{
				Name: "banana-network",
			}}
			filter = "banana"
		})

		It("lists, filters, and prompts for networks to delete", func() {
			list, err := networks.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListNetworksCall.CallCount).To(Equal(1))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("Network"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-network"))

			Expect(list).To(HaveLen(1))
		})

		Context("when the client fails to list networks", func() {
			BeforeEach(func() {
				client.ListNetworksCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := networks.List(filter)
				Expect(err).To(MatchError("List Networks: some error"))
			})
		})

		Context("when the network name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := networks.List("grape")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(0))
				Expect(list).To(HaveLen(0))
			})
		})

		Context("when it is the default network", func() {
			BeforeEach(func() {
				client.ListNetworksCall.Returns.Output = []*gcpcompute.Network{{
					Name: "default",
				}}
			})

			It("does not add it to the list", func() {
				list, err := networks.List("")
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
				list, err := networks.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})
})
