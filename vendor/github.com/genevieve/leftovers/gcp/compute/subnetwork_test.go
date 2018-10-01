package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Subnetwork", func() {
	var (
		client  *fakes.SubnetworksClient
		name    string
		region  string
		network string

		subnetwork compute.Subnetwork
	)

	BeforeEach(func() {
		client = &fakes.SubnetworksClient{}
		name = "banana"
		region = "region"
		network = "some-url/network"

		subnetwork = compute.NewSubnetwork(client, name, region, network)
	})

	Describe("Delete", func() {
		It("deletes the subnetwork", func() {
			err := subnetwork.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteSubnetworkCall.CallCount).To(Equal(1))
			Expect(client.DeleteSubnetworkCall.Receives.Subnetwork).To(Equal(name))
			Expect(client.DeleteSubnetworkCall.Receives.Region).To(Equal(region))
		})

		Context("when the client fails to delete because it was created by a network in auto subnet mode", func() {
			BeforeEach(func() {
				client.DeleteSubnetworkCall.Returns.Error = errors.New("Cannot delete auto subnetwork from an auto subnet mode network.")
			})

			It("returns success", func() {
				err := subnetwork.Delete()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the client fails to delete", func() {
			BeforeEach(func() {
				client.DeleteSubnetworkCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := subnetwork.Delete()
				Expect(err).To(MatchError("Delete: the-error"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(subnetwork.Name()).To(Equal("banana (Network:network)"))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(subnetwork.Type()).To(Equal("Subnetwork"))
		})
	})
})
