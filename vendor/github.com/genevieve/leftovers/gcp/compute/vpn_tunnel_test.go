package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("VpnTunnel", func() {
	var (
		client *fakes.VpnTunnelsClient
		name   string
		region string

		vpnTunnel compute.VpnTunnel
	)

	BeforeEach(func() {
		client = &fakes.VpnTunnelsClient{}
		name = "banana"
		region = "ca-cao"

		vpnTunnel = compute.NewVpnTunnel(client, name, region)
	})

	Describe("Delete", func() {
		It("deletes the vpn tunnel", func() {
			err := vpnTunnel.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteVpnTunnelCall.CallCount).To(Equal(1))
			Expect(client.DeleteVpnTunnelCall.Receives.VpnTunnel).To(Equal(name))
			Expect(client.DeleteVpnTunnelCall.Receives.Region).To(Equal(region))
		})

		Context("when the client fails to delete", func() {
			BeforeEach(func() {
				client.DeleteVpnTunnelCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := vpnTunnel.Delete()
				Expect(err).To(MatchError("Delete: the-error"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(vpnTunnel.Name()).To(Equal(name))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(vpnTunnel.Type()).To(Equal("Vpn Tunnel"))
		})
	})
})
