package ec2_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevievelesperance/leftovers/aws/ec2"
	"github.com/genevievelesperance/leftovers/aws/ec2/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NetworkInterface", func() {
	var (
		networkInterface ec2.NetworkInterface
		client           *fakes.NetworkInterfaceClient
		id               *string
	)

	BeforeEach(func() {
		client = &fakes.NetworkInterfaceClient{}
		id = aws.String("the-id")
		tags := []*awsec2.Tag{}

		networkInterface = ec2.NewNetworkInterface(client, id, tags)
	})

	It("deletes the network interface", func() {
		err := networkInterface.Delete()
		Expect(err).NotTo(HaveOccurred())

		Expect(client.DeleteNetworkInterfaceCall.CallCount).To(Equal(1))
		Expect(client.DeleteNetworkInterfaceCall.Receives.Input.NetworkInterfaceId).To(Equal(id))
	})

	Context("when the client fails", func() {
		BeforeEach(func() {
			client.DeleteNetworkInterfaceCall.Returns.Error = errors.New("banana")
		})

		It("returns the error", func() {
			err := networkInterface.Delete()
			Expect(err).To(MatchError("FAILED deleting network interface the-id: banana"))
		})
	})
})
