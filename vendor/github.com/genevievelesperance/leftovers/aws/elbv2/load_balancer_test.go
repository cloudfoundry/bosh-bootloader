package elbv2_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/genevievelesperance/leftovers/aws/elbv2"
	"github.com/genevievelesperance/leftovers/aws/elbv2/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LoadBalancer", func() {
	var (
		loadBalancer elbv2.LoadBalancer
		client       *fakes.LoadBalancersClient
		name         *string
		arn          *string
	)

	BeforeEach(func() {
		client = &fakes.LoadBalancersClient{}
		name = aws.String("the-name")
		arn = aws.String("the-arn")

		loadBalancer = elbv2.NewLoadBalancer(client, name, arn)
	})

	It("deletes the load balancer", func() {
		err := loadBalancer.Delete()
		Expect(err).NotTo(HaveOccurred())

		Expect(client.DeleteLoadBalancerCall.CallCount).To(Equal(1))
		Expect(client.DeleteLoadBalancerCall.Receives.Input.LoadBalancerArn).To(Equal(arn))
	})

	Context("when the client fails", func() {
		BeforeEach(func() {
			client.DeleteLoadBalancerCall.Returns.Error = errors.New("banana")
		})

		It("returns the error", func() {
			err := loadBalancer.Delete()
			Expect(err).To(MatchError("FAILED deleting load balancer the-name: banana"))
		})
	})
})
