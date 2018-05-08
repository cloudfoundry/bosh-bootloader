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

var _ = Describe("NatGateways", func() {
	var (
		client *fakes.NatGatewaysClient
		logger *fakes.Logger

		natGateways ec2.NatGateways
	)

	BeforeEach(func() {
		client = &fakes.NatGatewaysClient{}
		logger = &fakes.Logger{}

		natGateways = ec2.NewNatGateways(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.DescribeNatGatewaysCall.Returns.Output = &awsec2.DescribeNatGatewaysOutput{
				NatGateways: []*awsec2.NatGateway{{
					NatGatewayId: aws.String("banana"),
				}},
			}
			filter = "ban"
		})

		It("returns a list of resources to delete", func() {
			items, err := natGateways.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DescribeNatGatewaysCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("EC2 Nat Gateway"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana"))

			Expect(items).To(HaveLen(1))
		})

		Context("when the client fails to list resources", func() {
			BeforeEach(func() {
				client.DescribeNatGatewaysCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := natGateways.List(filter)
				Expect(err).To(MatchError("Describing EC2 Nat Gateways: some error"))
			})
		})

		Context("when the resource name does not contain the filter", func() {
			It("does not try deleting it", func() {
				items, err := natGateways.List("kiwi")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(0))
				Expect(items).To(HaveLen(0))
			})
		})

		Context("when the user responds no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("does not delete the resource", func() {
				items, err := natGateways.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
				Expect(items).To(HaveLen(0))
			})
		})
	})
})
