package ec2_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/genevievelesperance/leftovers/aws/ec2"
	"github.com/genevievelesperance/leftovers/aws/ec2/fakes"

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
			Expect(err).To(MatchError("FAILED releasing address the-public-ip: banana"))
		})
	})
})
