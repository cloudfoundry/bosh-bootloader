package elb_test

import (
	"errors"

	awselb "github.com/aws/aws-sdk-go/service/elb"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/elb"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Describer", func() {
	var (
		elbClient *fakes.ELBClient
		describer elb.Describer
	)

	BeforeEach(func() {
		elbClient = &fakes.ELBClient{}
		describer = elb.NewDescriber()
	})

	Describe("Describe", func() {
		It("describes the elb with the given name", func() {
			elbClient.DescribeLoadBalancersCall.Returns.Output = &awselb.DescribeLoadBalancersOutput{
				LoadBalancerDescriptions: []*awselb.LoadBalancerDescription{
					{
						Instances: []*awselb.Instance{
							{
								InstanceId: aws.String("some-instance-1"),
							},
							{
								InstanceId: aws.String("some-instance-2"),
							},
						},
					},
				},
			}

			instances, err := describer.Describe("some-elb-name", elbClient)
			Expect(err).NotTo(HaveOccurred())

			Expect(*elbClient.DescribeLoadBalancersCall.Receives.Input).To(Equal(
				awselb.DescribeLoadBalancersInput{
					LoadBalancerNames: []*string{
						aws.String("some-elb-name"),
					},
				},
			))
			Expect(instances).To(Equal([]string{"some-instance-1", "some-instance-2"}))
		})

		Context("failure cases", func() {
			It("returns an error when the load balancer cannot be described", func() {
				elbClient.DescribeLoadBalancersCall.Returns.Error = errors.New("something bad happened")

				_, err := describer.Describe("some-elb-name", elbClient)
				Expect(err).To(MatchError("something bad happened"))
			})
		})
	})
})
