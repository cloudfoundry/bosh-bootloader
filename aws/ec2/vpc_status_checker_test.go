package ec2_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"

	"github.com/cloudfoundry/bosh-bootloader/aws/ec2"
	"github.com/cloudfoundry/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("VPCStatusChecker", func() {
	var (
		vpcStatusChecker ec2.VPCStatusChecker
		ec2Client        *fakes.EC2Client
		clientProvider   *fakes.ClientProvider
	)

	BeforeEach(func() {
		clientProvider = &fakes.ClientProvider{}
		ec2Client = &fakes.EC2Client{}
		clientProvider.GetEC2ClientCall.Returns.EC2Client = ec2Client
		vpcStatusChecker = ec2.NewVPCStatusChecker(clientProvider)
	})

	Describe("ValidateSafeToDelete", func() {
		var reservationContainingInstance = func(tag string) *awsec2.Reservation {
			return &awsec2.Reservation{
				Instances: []*awsec2.Instance{{
					Tags: []*awsec2.Tag{{
						Key:   aws.String("Name"),
						Value: aws.String(tag),
					}},
				}},
			}
		}

		It("returns nil when the only EC2 instances are bosh and nat", func() {
			ec2Client.DescribeInstancesCall.Returns.Output = &awsec2.DescribeInstancesOutput{
				Reservations: []*awsec2.Reservation{
					reservationContainingInstance("NAT"),
					reservationContainingInstance("bosh/0"),
				},
			}

			err := vpcStatusChecker.ValidateSafeToDelete("some-vpc-id")
			Expect(err).NotTo(HaveOccurred())

			Expect(ec2Client.DescribeInstancesCall.Receives.Input).To(Equal(&awsec2.DescribeInstancesInput{
				Filters: []*awsec2.Filter{{
					Name:   aws.String("vpc-id"),
					Values: []*string{aws.String("some-vpc-id")},
				}},
			}))
		})

		It("returns nil when there are no instances at all", func() {
			ec2Client.DescribeInstancesCall.Returns.Output = &awsec2.DescribeInstancesOutput{
				Reservations: []*awsec2.Reservation{},
			}

			err := vpcStatusChecker.ValidateSafeToDelete("some-vpc-id")
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns an error when there are bosh-deployed VMs in the VPC", func() {
			ec2Client.DescribeInstancesCall.Returns.Output = &awsec2.DescribeInstancesOutput{
				Reservations: []*awsec2.Reservation{
					reservationContainingInstance("NAT"),
					reservationContainingInstance("bosh/0"),
					reservationContainingInstance("first-bosh-deployed-vm"),
					reservationContainingInstance("second-bosh-deployed-vm"),
				},
			}

			err := vpcStatusChecker.ValidateSafeToDelete("some-vpc-id")
			Expect(err).To(MatchError("vpc some-vpc-id is not safe to delete; vms still exist: [first-bosh-deployed-vm, second-bosh-deployed-vm]"))
		})

		It("returns an error even when there are two VMs in the VPC, but they are not NAT and BOSH", func() {
			ec2Client.DescribeInstancesCall.Returns.Output = &awsec2.DescribeInstancesOutput{
				Reservations: []*awsec2.Reservation{
					reservationContainingInstance("not-bosh"),
					reservationContainingInstance("not-nat"),
				},
			}

			err := vpcStatusChecker.ValidateSafeToDelete("some-vpc-id")
			Expect(err).To(MatchError("vpc some-vpc-id is not safe to delete; vms still exist: [not-bosh, not-nat]"))
		})

		It("returns an error even if the vpc contains other instances tagged NAT and bosh/0", func() {
			ec2Client.DescribeInstancesCall.Returns.Output = &awsec2.DescribeInstancesOutput{
				Reservations: []*awsec2.Reservation{
					reservationContainingInstance("NAT"),
					reservationContainingInstance("NAT"),
					reservationContainingInstance("bosh/0"),
					reservationContainingInstance("bosh/0"),
					reservationContainingInstance("bosh/0"),
				},
			}

			err := vpcStatusChecker.ValidateSafeToDelete("some-vpc-id")
			Expect(err).To(MatchError("vpc some-vpc-id is not safe to delete; vms still exist: [NAT, bosh/0, bosh/0]"))
		})

		It("returns an error even if the vpc contains untagged vms", func() {
			ec2Client.DescribeInstancesCall.Returns.Output = &awsec2.DescribeInstancesOutput{
				Reservations: []*awsec2.Reservation{
					&awsec2.Reservation{
						Instances: []*awsec2.Instance{{
							Tags: []*awsec2.Tag{{
								Key:   aws.String("Name"),
								Value: aws.String(""),
							}},
						}, {}, {}},
					},
				},
			}

			err := vpcStatusChecker.ValidateSafeToDelete("some-vpc-id")
			Expect(err).To(MatchError("vpc some-vpc-id is not safe to delete; vms still exist: [unnamed, unnamed, unnamed]"))
		})

		Describe("failure cases", func() {
			It("returns an error when the describe instances call fails", func() {
				ec2Client.DescribeInstancesCall.Returns.Error = errors.New("failed to describe instances")
				err := vpcStatusChecker.ValidateSafeToDelete("some-vpc-id")
				Expect(err).To(MatchError("failed to describe instances"))
			})
		})
	})
})
