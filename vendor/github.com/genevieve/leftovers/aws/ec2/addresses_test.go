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

var _ = Describe("Addresses", func() {
	var (
		client *fakes.AddressesClient
		logger *fakes.Logger

		addresses ec2.Addresses
	)

	BeforeEach(func() {
		client = &fakes.AddressesClient{}
		logger = &fakes.Logger{}

		addresses = ec2.NewAddresses(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.DescribeAddressesCall.Returns.Output = &awsec2.DescribeAddressesOutput{
				Addresses: []*awsec2.Address{{
					PublicIp:     aws.String("banana"),
					AllocationId: aws.String("the-allocation-id"),
					InstanceId:   aws.String(""),
				}},
			}
			filter = "ban"
		})

		It("releases ec2 addresses", func() {
			items, err := addresses.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DescribeAddressesCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("EC2 Address"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana"))

			Expect(items).To(HaveLen(1))
		})

		Context("when the address tags do not contain the filter", func() {
			It("does not try releasing them", func() {
				_, err := addresses.List("kiwi")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(0))
			})
		})

		Context("when the client fails to describe addresses", func() {
			BeforeEach(func() {
				client.DescribeAddressesCall.Returns.Error = errors.New("some error")
			})

			It("does not try releasing them", func() {
				_, err := addresses.List(filter)
				Expect(err).To(MatchError("Describing EC2 Addresses: some error"))
			})
		})

		Context("when the user responds no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("does not release the address", func() {
				items, err := addresses.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
				Expect(items).To(HaveLen(0))
			})
		})
	})
})
