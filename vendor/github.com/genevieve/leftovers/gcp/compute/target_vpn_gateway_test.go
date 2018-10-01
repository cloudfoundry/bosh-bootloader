package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TargetVpnGateway", func() {
	var (
		client *fakes.TargetVpnGatewaysClient
		name   string
		region string

		targetVpnGateway compute.TargetVpnGateway
	)

	BeforeEach(func() {
		client = &fakes.TargetVpnGatewaysClient{}
		name = "banana"
		region = "region"

		targetVpnGateway = compute.NewTargetVpnGateway(client, name, region)
	})

	Describe("Delete", func() {
		It("deletes the resource", func() {
			err := targetVpnGateway.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteTargetVpnGatewayCall.CallCount).To(Equal(1))
			Expect(client.DeleteTargetVpnGatewayCall.Receives.TargetVpnGateway).To(Equal(name))
		})

		Context("when the client fails to delete", func() {
			BeforeEach(func() {
				client.DeleteTargetVpnGatewayCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := targetVpnGateway.Delete()
				Expect(err).To(MatchError("Delete: the-error"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(targetVpnGateway.Name()).To(Equal(name))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(targetVpnGateway.Type()).To(Equal("Target Vpn Gateway"))
		})
	})
})
