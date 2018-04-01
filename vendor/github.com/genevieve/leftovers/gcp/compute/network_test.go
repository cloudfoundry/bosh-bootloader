package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Network", func() {
	var (
		client *fakes.NetworksClient
		name   string

		network compute.Network
	)

	BeforeEach(func() {
		client = &fakes.NetworksClient{}
		name = "banana"

		network = compute.NewNetwork(client, name)
	})

	Describe("Delete", func() {
		It("deletes the network", func() {
			err := network.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteNetworkCall.CallCount).To(Equal(1))
			Expect(client.DeleteNetworkCall.Receives.Network).To(Equal(name))
		})

		Context("when the client fails to delete", func() {
			BeforeEach(func() {
				client.DeleteNetworkCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := network.Delete()
				Expect(err).To(MatchError("Delete: the-error"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(network.Name()).To(Equal(name))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(network.Type()).To(Equal("Network"))
		})
	})
})
