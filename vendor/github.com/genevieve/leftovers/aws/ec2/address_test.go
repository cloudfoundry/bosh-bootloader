package ec2_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/genevieve/leftovers/aws/ec2"
	"github.com/genevieve/leftovers/aws/ec2/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Address", func() {
	var (
		address      ec2.Address
		client       *fakes.AddressesClient
		publicIp     *string
		allocationId *string
	)

	BeforeEach(func() {
		client = &fakes.AddressesClient{}
		publicIp = aws.String("the-public-ip")
		allocationId = aws.String("the-allocation-id")

		address = ec2.NewAddress(client, publicIp, allocationId)
	})

	Describe("Delete", func() {
		It("releases the address", func() {
			err := address.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ReleaseAddressCall.CallCount).To(Equal(1))
			Expect(client.ReleaseAddressCall.Receives.Input.AllocationId).To(Equal(allocationId))
		})

		Context("the client fails", func() {
			BeforeEach(func() {
				client.ReleaseAddressCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := address.Delete()
				Expect(err).To(MatchError("Delete: banana"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the identifier", func() {
			Expect(address.Name()).To(Equal("the-public-ip"))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(address.Type()).To(Equal("EC2 Address"))
		})
	})
})
