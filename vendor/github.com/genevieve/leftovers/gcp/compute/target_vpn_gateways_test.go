package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gcpcompute "google.golang.org/api/compute/v1"
)

var _ = Describe("TargetVpnGateways", func() {
	var (
		client  *fakes.TargetVpnGatewaysClient
		logger  *fakes.Logger
		regions map[string]string

		targetVpnGateways compute.TargetVpnGateways
	)

	BeforeEach(func() {
		client = &fakes.TargetVpnGatewaysClient{}
		logger = &fakes.Logger{}
		regions = map[string]string{"https://region-1": "region-1"}

		targetVpnGateways = compute.NewTargetVpnGateways(client, logger, regions)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.ListTargetVpnGatewaysCall.Returns.Output = []*gcpcompute.TargetVpnGateway{
				{
					Name:   "banana-target-vpn-gateway",
					Region: "https://region-1",
				},
			}
			filter = "banana"
		})

		It("lists, filters, and prompts for target vpn gateways to delete", func() {
			list, err := targetVpnGateways.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListTargetVpnGatewaysCall.CallCount).To(Equal(1))
			Expect(client.ListTargetVpnGatewaysCall.Receives.Region).To(Equal("region-1"))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("Target Vpn Gateway"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-target-vpn-gateway"))

			Expect(list).To(HaveLen(1))
		})

		Context("when the client fails to list target vpn gateways", func() {
			BeforeEach(func() {
				client.ListTargetVpnGatewaysCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := targetVpnGateways.List(filter)
				Expect(err).To(MatchError("List Target Vpn Gateways: some error"))
			})
		})

		Context("when the target vpn gateway name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := targetVpnGateways.List("grape")
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
				list, err := targetVpnGateways.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})
})
