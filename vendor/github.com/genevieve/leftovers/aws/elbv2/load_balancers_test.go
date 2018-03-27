package elbv2_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	awselbv2 "github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/genevieve/leftovers/aws/elbv2"
	"github.com/genevieve/leftovers/aws/elbv2/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LoadBalancers", func() {
	var (
		client *fakes.LoadBalancersClient
		logger *fakes.Logger

		loadBalancers elbv2.LoadBalancers
	)

	BeforeEach(func() {
		client = &fakes.LoadBalancersClient{}
		logger = &fakes.Logger{}

		loadBalancers = elbv2.NewLoadBalancers(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.DescribeLoadBalancersCall.Returns.Output = &awselbv2.DescribeLoadBalancersOutput{
				LoadBalancers: []*awselbv2.LoadBalancer{{
					LoadBalancerName: aws.String("banana"),
					LoadBalancerArn:  aws.String("the-arn"),
				}},
			}
			filter = "banana"
		})

		It("returns a list of elbv2 load balancers to delete", func() {
			items, err := loadBalancers.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DescribeLoadBalancersCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("load balancer"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana"))

			Expect(items).To(HaveLen(1))
		})

		Context("when the client fails to list load balancers", func() {
			BeforeEach(func() {
				client.DescribeLoadBalancersCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := loadBalancers.List(filter)
				Expect(err).To(MatchError("Describing load balancers: some error"))
			})
		})

		Context("when the load balancer name does not contain the filter", func() {
			It("does not return it in the list", func() {
				items, err := loadBalancers.List("kiwi")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(0))
				Expect(items).To(HaveLen(0))
			})
		})

		Context("when the user responds no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("does not return it in the list", func() {
				items, err := loadBalancers.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
				Expect(items).To(HaveLen(0))
			})
		})
	})
})
