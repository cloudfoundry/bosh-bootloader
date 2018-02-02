package compute_test

import (
	"errors"

	"github.com/genevievelesperance/leftovers/gcp/compute"
	"github.com/genevievelesperance/leftovers/gcp/compute/fakes"
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
			logger.PromptCall.Returns.Proceed = true
			client.ListNetworksCall.Returns.Output = &gcpcompute.NetworkList{
				Items: []*gcpcompute.Network{{
					Name: "banana-network",
				}},
			}
			filter = "banana"
		})

		It("lists, filters, and prompts for networks to delete", func() {
			list, err := networks.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListNetworksCall.CallCount).To(Equal(1))

			Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to delete network banana-network?"))

			Expect(list).To(HaveLen(1))
			Expect(list).To(HaveKeyWithValue("banana-network", ""))
		})

		Context("when the client fails to list networks", func() {
			BeforeEach(func() {
				client.ListNetworksCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := networks.List(filter)
				Expect(err).To(MatchError("Listing networks: some error"))
			})
		})

		Context("when the network name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := networks.List("grape")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.CallCount).To(Equal(0))
				Expect(list).To(HaveLen(0))
			})
		})

		Context("when it is the default network", func() {
			BeforeEach(func() {
				client.ListNetworksCall.Returns.Output = &gcpcompute.NetworkList{
					Items: []*gcpcompute.Network{{
						Name: "default",
					}},
				}
			})

			It("does not add it to the list", func() {
				list, err := networks.List(filter)
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
				list, err := networks.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})

	Describe("Delete", func() {
		var list map[string]string

		BeforeEach(func() {
			list = map[string]string{"banana-network": ""}
		})

		It("deletes networks", func() {
			networks.Delete(list)

			Expect(client.DeleteNetworkCall.CallCount).To(Equal(1))
			Expect(client.DeleteNetworkCall.Receives.Network).To(Equal("banana-network"))

			Expect(logger.PrintfCall.Messages).To(Equal([]string{"SUCCESS deleting network banana-network\n"}))
		})

		Context("when the client fails to delete a network", func() {
			BeforeEach(func() {
				client.DeleteNetworkCall.Returns.Error = errors.New("some error")
			})

			It("logs the error", func() {
				networks.Delete(list)

				Expect(logger.PrintfCall.Messages).To(Equal([]string{"ERROR deleting network banana-network: some error\n"}))
			})
		})
	})
})
