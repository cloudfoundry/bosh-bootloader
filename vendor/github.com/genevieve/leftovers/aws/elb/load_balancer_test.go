package elb_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/genevieve/leftovers/aws/elb"
	"github.com/genevieve/leftovers/aws/elb/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LoadBalancer", func() {
	var (
		loadBalancer elb.LoadBalancer
		client       *fakes.LoadBalancersClient
		name         *string
	)

	BeforeEach(func() {
		client = &fakes.LoadBalancersClient{}
		name = aws.String("the-name")

		loadBalancer = elb.NewLoadBalancer(client, name)
	})

	Describe("Delete", func() {
		It("deletes the load balancer", func() {
			err := loadBalancer.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteLoadBalancerCall.CallCount).To(Equal(1))
			Expect(client.DeleteLoadBalancerCall.Receives.Input.LoadBalancerName).To(Equal(name))
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

	Describe("Name", func() {
		It("returns the identifier", func() {
			Expect(loadBalancer.Name()).To(Equal("the-name"))
		})
	})
})
