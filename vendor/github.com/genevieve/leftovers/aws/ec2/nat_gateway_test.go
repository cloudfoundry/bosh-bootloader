package ec2_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevieve/leftovers/aws/ec2"
	"github.com/genevieve/leftovers/aws/ec2/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NatGateway", func() {
	var (
		natGateway ec2.NatGateway
		client     *fakes.NatGatewaysClient
		logger     *fakes.Logger
		id         *string
	)

	BeforeEach(func() {
		client = &fakes.NatGatewaysClient{}
		logger = &fakes.Logger{}
		id = aws.String("the-id")
		tags := []*awsec2.Tag{{Key: aws.String("the-key"), Value: aws.String("the-value")}}

		natGateway = ec2.NewNatGateway(client, logger, id, tags)
	})

	Describe("Delete", func() {
		BeforeEach(func() {
			client.DescribeNatGatewaysCall.Returns.Output = &awsec2.DescribeNatGatewaysOutput{
				NatGateways: []*awsec2.NatGateway{{
					NatGatewayId: id,
					State:        aws.String("deleted"),
				}},
			}
		})
		It("deletes the resource", func() {
			err := natGateway.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteNatGatewayCall.CallCount).To(Equal(1))
			Expect(client.DeleteNatGatewayCall.Receives.Input.NatGatewayId).To(Equal(id))
		})

		Context("when the client fails", func() {
			BeforeEach(func() {
				client.DeleteNatGatewayCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := natGateway.Delete()
				Expect(err).To(MatchError("Delete: banana"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the identifier", func() {
			Expect(natGateway.Name()).To(Equal("the-id (the-key:the-value)"))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(natGateway.Type()).To(Equal("EC2 Nat Gateway"))
		})
	})
})
