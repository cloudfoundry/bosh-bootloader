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

var _ = Describe("InternetGateways", func() {
	var (
		client *fakes.InternetGatewaysClient
		logger *fakes.Logger

		gateways ec2.InternetGateways
	)

	BeforeEach(func() {
		client = &fakes.InternetGatewaysClient{}
		logger = &fakes.Logger{}

		gateways = ec2.NewInternetGateways(client, logger)
	})

	Describe("Delete", func() {
		BeforeEach(func() {
			client.DescribeInternetGatewaysCall.Returns.Output = &awsec2.DescribeInternetGatewaysOutput{
				InternetGateways: []*awsec2.InternetGateway{{
					InternetGatewayId: aws.String("the-gateway-id"),
					Attachments: []*awsec2.InternetGatewayAttachment{{
						VpcId: aws.String("the-vpc-id"),
					}},
				}},
			}
		})

		It("detaches and deletes the internet gateways", func() {
			err := gateways.Delete("the-vpc-id")
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DescribeInternetGatewaysCall.CallCount).To(Equal(1))
			Expect(client.DescribeInternetGatewaysCall.Receives.Input.Filters[0].Name).To(Equal(aws.String("attachment.vpc-id")))
			Expect(client.DescribeInternetGatewaysCall.Receives.Input.Filters[0].Values[0]).To(Equal(aws.String("the-vpc-id")))

			Expect(client.DetachInternetGatewayCall.CallCount).To(Equal(1))
			Expect(client.DetachInternetGatewayCall.Receives.Input.InternetGatewayId).To(Equal(aws.String("the-gateway-id")))
			Expect(client.DetachInternetGatewayCall.Receives.Input.VpcId).To(Equal(aws.String("the-vpc-id")))

			Expect(client.DeleteInternetGatewayCall.CallCount).To(Equal(1))
			Expect(client.DeleteInternetGatewayCall.Receives.Input.InternetGatewayId).To(Equal(aws.String("the-gateway-id")))

			Expect(logger.PrintfCall.Messages).To(Equal([]string{
				"SUCCESS detaching internet gateway the-gateway-id\n",
				"SUCCESS deleting internet gateway the-gateway-id\n",
			}))
		})

		Context("when the client fails to describe attached internet gateways", func() {
			BeforeEach(func() {
				client.DescribeInternetGatewaysCall.Returns.Error = errors.New("some error")
			})

			It("returns the error and does not try deleting them", func() {
				err := gateways.Delete("banana")
				Expect(err).To(MatchError("Describing internet gateways: some error"))

				Expect(client.DetachInternetGatewayCall.CallCount).To(Equal(0))
				Expect(client.DeleteInternetGatewayCall.CallCount).To(Equal(0))
			})
		})

		Context("when the client fails to detach the internet gateway", func() {
			BeforeEach(func() {
				client.DetachInternetGatewayCall.Returns.Error = errors.New("some error")
			})

			It("logs the error and deletes the internet gateway", func() {
				err := gateways.Delete("banana")
				Expect(err).NotTo(HaveOccurred())

				Expect(client.DeleteInternetGatewayCall.CallCount).To(Equal(1))
				Expect(logger.PrintfCall.Messages).To(Equal([]string{
					"ERROR detaching internet gateway the-gateway-id: some error\n",
					"SUCCESS deleting internet gateway the-gateway-id\n",
				}))
			})
		})

		Context("when the client fails to delete the internet gateway", func() {
			BeforeEach(func() {
				client.DeleteInternetGatewayCall.Returns.Error = errors.New("some error")
			})

			It("logs the error", func() {
				err := gateways.Delete("banana")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PrintfCall.Messages).To(Equal([]string{
					"SUCCESS detaching internet gateway the-gateway-id\n",
					"ERROR deleting internet gateway the-gateway-id: some error\n",
				}))
			})
		})
	})
})
