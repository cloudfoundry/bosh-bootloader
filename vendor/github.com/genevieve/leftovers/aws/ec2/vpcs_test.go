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

var _ = Describe("Vpcs", func() {
	var (
		client *fakes.VpcClient
		logger *fakes.Logger

		vpcs ec2.Vpcs
	)

	BeforeEach(func() {
		client = &fakes.VpcClient{}
		logger = &fakes.Logger{}
		routes := &fakes.RouteTables{}
		subnets := &fakes.Subnets{}
		gateways := &fakes.InternetGateways{}

		vpcs = ec2.NewVpcs(client, logger, routes, subnets, gateways)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.DescribeVpcsCall.Returns.Output = &awsec2.DescribeVpcsOutput{
				Vpcs: []*awsec2.Vpc{{
					IsDefault: aws.Bool(false),
					Tags: []*awsec2.Tag{{
						Key:   aws.String("Name"),
						Value: aws.String("banana"),
					}},
					VpcId: aws.String("the-vpc-id"),
				}},
			}
			filter = "ban"
		})

		It("returns a list of vpcs to delete", func() {
			items, err := vpcs.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DescribeVpcsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("vpc"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("the-vpc-id (Name:banana)"))

			Expect(items).To(HaveLen(1))
		})

		Context("when the vpc tags contain the filter", func() {
			It("does not return it in the list", func() {
				items, err := vpcs.List("kiwi")
				Expect(err).NotTo(HaveOccurred())

				Expect(items).To(HaveLen(0))
			})
		})

		Context("when the vpc is a default", func() {
			BeforeEach(func() {
				client.DescribeVpcsCall.Returns.Output = &awsec2.DescribeVpcsOutput{
					Vpcs: []*awsec2.Vpc{{
						IsDefault: aws.Bool(true),
						VpcId:     aws.String("the-vpc-id"),
					}},
				}
			})

			It("does not return it in the list", func() {
				items, err := vpcs.List(filter)
				Expect(err).NotTo(HaveOccurred())
				Expect(items).To(HaveLen(0))
			})
		})

		Context("when there is no tag name", func() {
			BeforeEach(func() {
				client.DescribeVpcsCall.Returns.Output = &awsec2.DescribeVpcsOutput{
					Vpcs: []*awsec2.Vpc{{
						IsDefault: aws.Bool(false),
						VpcId:     aws.String("the-vpc-id"),
					}},
				}
			})

			It("uses just the vpc id in the prompt", func() {
				items, err := vpcs.List("the-vpc")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("the-vpc-id"))
				Expect(items).To(HaveLen(1))
			})
		})

		Context("when the client fails to list vpcs", func() {
			BeforeEach(func() {
				client.DescribeVpcsCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := vpcs.List(filter)
				Expect(err).To(MatchError("Describing vpcs: some error"))
			})
		})

		Context("when the user responds no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("does not return it in the list", func() {
				items, err := vpcs.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
				Expect(items).To(HaveLen(0))
			})
		})
	})
})
