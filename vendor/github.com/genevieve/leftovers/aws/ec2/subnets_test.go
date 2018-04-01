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

var _ = Describe("Subnets", func() {
	var (
		client       *fakes.SubnetsClient
		logger       *fakes.Logger
		resourceTags *fakes.ResourceTags

		subnets ec2.Subnets
	)

	BeforeEach(func() {
		client = &fakes.SubnetsClient{}
		logger = &fakes.Logger{}
		resourceTags = &fakes.ResourceTags{}

		subnets = ec2.NewSubnets(client, logger, resourceTags)
	})

	Describe("Delete", func() {
		BeforeEach(func() {
			client.DescribeSubnetsCall.Returns.Output = &awsec2.DescribeSubnetsOutput{
				Subnets: []*awsec2.Subnet{{
					SubnetId: aws.String("the-subnet-id"),
					VpcId:    aws.String("the-vpc-id"),
				}},
			}
		})

		It("deletes the subnets", func() {
			err := subnets.Delete("the-vpc-id")
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DescribeSubnetsCall.CallCount).To(Equal(1))
			Expect(client.DescribeSubnetsCall.Receives.Input.Filters[0].Name).To(Equal(aws.String("vpc-id")))
			Expect(client.DescribeSubnetsCall.Receives.Input.Filters[0].Values[0]).To(Equal(aws.String("the-vpc-id")))

			Expect(client.DeleteSubnetCall.CallCount).To(Equal(1))
			Expect(client.DeleteSubnetCall.Receives.Input.SubnetId).To(Equal(aws.String("the-subnet-id")))

			Expect(resourceTags.DeleteCall.CallCount).To(Equal(1))
			Expect(resourceTags.DeleteCall.Receives.ResourceType).To(Equal("subnet"))
			Expect(resourceTags.DeleteCall.Receives.ResourceId).To(Equal("the-subnet-id"))

			Expect(logger.PrintfCall.Messages).To(Equal([]string{
				"[EC2 VPC: the-vpc-id] Deleted subnet the-subnet-id tags",
			}))
		})

		Context("when the client fails to describe subnets", func() {
			BeforeEach(func() {
				client.DescribeSubnetsCall.Returns.Error = errors.New("some error")
			})

			It("returns the error and does not try deleting them", func() {
				err := subnets.Delete("banana")
				Expect(err).To(MatchError("Describe EC2 Subnets: some error"))

				Expect(client.DeleteSubnetCall.CallCount).To(Equal(0))
			})
		})

		Context("when the client fails to delete the subnet", func() {
			BeforeEach(func() {
				client.DeleteSubnetCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				err := subnets.Delete("banana")
				Expect(err).To(MatchError("Delete subnet the-subnet-id: some error"))
			})
		})

		Context("when the resource tags fails to delete", func() {
			BeforeEach(func() {
				resourceTags.DeleteCall.Returns.Error = errors.New("some error")
			})

			It("logs the error", func() {
				err := subnets.Delete("banana")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PrintfCall.Messages).To(Equal([]string{
					"[EC2 VPC: banana] Delete subnet the-subnet-id tags: some error",
				}))
			})
		})
	})
})
