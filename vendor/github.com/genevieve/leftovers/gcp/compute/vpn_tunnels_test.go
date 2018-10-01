package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gcpcompute "google.golang.org/api/compute/v1"
)

var _ = Describe("VpnTunnels", func() {
	var (
		client  *fakes.VpnTunnelsClient
		logger  *fakes.Logger
		regions map[string]string

		vpnTunnels compute.VpnTunnels
	)

	BeforeEach(func() {
		client = &fakes.VpnTunnelsClient{}
		logger = &fakes.Logger{}
		regions = map[string]string{"https://region-1": "region-1"}

		vpnTunnels = compute.NewVpnTunnels(client, logger, regions)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.ListVpnTunnelsCall.Returns.Output = []*gcpcompute.VpnTunnel{
				{
					Name:   "banana-vpn-tunnel",
					Region: "https://region-1",
				},
			}
			filter = "banana"
		})

		It("lists, filters, and prompts for vpn tunnels to delete", func() {
			list, err := vpnTunnels.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListVpnTunnelsCall.CallCount).To(Equal(1))
			Expect(client.ListVpnTunnelsCall.Receives.Region).To(Equal("region-1"))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("Vpn Tunnel"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-vpn-tunnel"))

			Expect(list).To(HaveLen(1))
		})

		Context("when the client fails to list vpn tunnels", func() {
			BeforeEach(func() {
				client.ListVpnTunnelsCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := vpnTunnels.List(filter)
				Expect(err).To(MatchError("List Vpn Tunnels: some error"))
			})
		})

		Context("when the vpn tunnel name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := vpnTunnels.List("grape")
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
				list, err := vpnTunnels.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})
})
