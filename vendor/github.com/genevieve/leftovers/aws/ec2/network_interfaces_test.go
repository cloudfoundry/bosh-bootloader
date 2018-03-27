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

var _ = Describe("NetworkInterfaces", func() {
	var (
		client *fakes.NetworkInterfaceClient
		logger *fakes.Logger

		networkInterfaces ec2.NetworkInterfaces
	)

	BeforeEach(func() {
		client = &fakes.NetworkInterfaceClient{}
		logger = &fakes.Logger{}

		networkInterfaces = ec2.NewNetworkInterfaces(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.DescribeNetworkInterfacesCall.Returns.Output = &awsec2.DescribeNetworkInterfacesOutput{
				NetworkInterfaces: []*awsec2.NetworkInterface{{
					NetworkInterfaceId: aws.String("banana"),
				}},
			}
			filter = "ban"
		})

		It("returns a list of network interfaces to delete", func() {
			items, err := networkInterfaces.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DescribeNetworkInterfacesCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("network interface"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana"))

			Expect(items).To(HaveLen(1))
		})

		Context("when the client fails to list network interfaces", func() {
			BeforeEach(func() {
				client.DescribeNetworkInterfacesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := networkInterfaces.List(filter)
				Expect(err).To(MatchError("Describing network interfaces: some error"))
			})
		})

		Context("when the network interface name does not contain the filter", func() {
			It("does not return it in the list", func() {
				items, err := networkInterfaces.List("kiwi")
				Expect(err).NotTo(HaveOccurred())

				Expect(client.DescribeNetworkInterfacesCall.CallCount).To(Equal(1))
				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(0))
				Expect(items).To(HaveLen(0))
			})
		})

		Context("when the network interface has tags", func() {
			BeforeEach(func() {
				client.DescribeNetworkInterfacesCall.Returns.Output = &awsec2.DescribeNetworkInterfacesOutput{
					NetworkInterfaces: []*awsec2.NetworkInterface{{
						NetworkInterfaceId: aws.String("banana"),
						TagSet: []*awsec2.Tag{{
							Key:   aws.String("the-key"),
							Value: aws.String("the-value"),
						}},
					}},
				}
			})

			It("uses them in the prompt", func() {
				_, err := networkInterfaces.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana (the-key:the-value)"))
			})
		})

		Context("when the user responds no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("does not return it in the list", func() {
				items, err := networkInterfaces.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
				Expect(items).To(HaveLen(0))
			})
		})
	})
})
