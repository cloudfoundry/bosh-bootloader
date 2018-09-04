package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gcpcompute "google.golang.org/api/compute/v1"
)

var _ = Describe("Subnetworks", func() {
	var (
		client  *fakes.SubnetworksClient
		logger  *fakes.Logger
		regions map[string]string

		subnetworks compute.Subnetworks
	)

	BeforeEach(func() {
		client = &fakes.SubnetworksClient{}
		logger = &fakes.Logger{}
		regions = map[string]string{"https://region-1": "region-1"}

		subnetworks = compute.NewSubnetworks(client, logger, regions)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.ListSubnetworksCall.Returns.Output = []*gcpcompute.Subnetwork{{
				Name:   "banana-subnetwork",
				Region: "https://region-1",
			}}
			filter = "banana"
		})

		It("lists, filters, and prompts for subnetworks to delete", func() {
			list, err := subnetworks.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListSubnetworksCall.CallCount).To(Equal(1))
			Expect(client.ListSubnetworksCall.Receives.Region).To(Equal("region-1"))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("Subnetwork"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-subnetwork"))

			Expect(list).To(HaveLen(1))
		})

		Context("when the client fails to list subnetworks", func() {
			BeforeEach(func() {
				client.ListSubnetworksCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := subnetworks.List(filter)
				Expect(err).To(MatchError("List Subnetworks for region region-1: some error"))
			})
		})

		Context("when the subnetwork name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := subnetworks.List("grape")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(0))
				Expect(list).To(HaveLen(0))
			})
		})

		Context("when it is the default subnetwork", func() {
			BeforeEach(func() {
				client.ListSubnetworksCall.Returns.Output = []*gcpcompute.Subnetwork{{
					Name:   "default",
					Region: "https://region-1",
				}}
			})

			It("does not add it to the list", func() {
				list, err := subnetworks.List("")
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
				list, err := subnetworks.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})
})
