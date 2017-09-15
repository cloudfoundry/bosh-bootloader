package ec2_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/cloudfoundry/bosh-bootloader/aws"
	"github.com/cloudfoundry/bosh-bootloader/aws/ec2"
	"github.com/cloudfoundry/bosh-bootloader/fakes"

	awslib "github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	Describe("NewClient", func() {
		It("returns a Client with the provided configuration", func() {
			client := ec2.NewClient(
				aws.Config{
					AccessKeyID:     "some-access-key-id",
					SecretAccessKey: "some-secret-access-key",
					Region:          "some-region",
				},
				&fakes.Logger{},
			)

			ec2Client, ok := client.GetEC2Client().(*awsec2.EC2)
			Expect(ok).To(BeTrue())

			Expect(ec2Client.Config.Credentials).To(Equal(credentials.NewStaticCredentials("some-access-key-id", "some-secret-access-key", "")))
			Expect(ec2Client.Config.Region).To(Equal(awslib.String("some-region")))
		})
	})

	Describe("RetrieveAvailabilityZones", func() {
		var (
			client    ec2.Client
			ec2Client *fakes.AWSEC2Client
		)

		BeforeEach(func() {
			ec2Client = &fakes.AWSEC2Client{}
			client = ec2.NewClientWithInjectedEC2Client(ec2Client, &fakes.Logger{})
		})

		It("fetches availability zones for a given region", func() {
			ec2Client.DescribeAvailabilityZonesCall.Returns.Output = &awsec2.DescribeAvailabilityZonesOutput{
				AvailabilityZones: []*awsec2.AvailabilityZone{
					{ZoneName: awslib.String("us-east-1a")},
					{ZoneName: awslib.String("us-east-1b")},
					{ZoneName: awslib.String("us-east-1c")},
					{ZoneName: awslib.String("us-east-1e")},
				},
			}

			azs, err := client.RetrieveAvailabilityZones("us-east-1")

			Expect(err).NotTo(HaveOccurred())
			Expect(azs).To(ConsistOf("us-east-1a", "us-east-1b", "us-east-1c", "us-east-1e"))
			Expect(ec2Client.DescribeAvailabilityZonesCall.Receives.Input).To(Equal(&awsec2.DescribeAvailabilityZonesInput{
				Filters: []*awsec2.Filter{{
					Name:   awslib.String("region-name"),
					Values: []*string{awslib.String("us-east-1")},
				}},
			}))
		})

		Describe("failure cases", func() {
			It("returns an error when AWS returns a nil availability zone", func() {
				ec2Client.DescribeAvailabilityZonesCall.Returns.Output = &awsec2.DescribeAvailabilityZonesOutput{
					AvailabilityZones: []*awsec2.AvailabilityZone{nil},
				}

				_, err := client.RetrieveAvailabilityZones("us-east-1")
				Expect(err).To(MatchError("aws returned nil availability zone"))
			})

			It("returns an error when an availability zone with a nil ZoneName", func() {
				ec2Client.DescribeAvailabilityZonesCall.Returns.Output = &awsec2.DescribeAvailabilityZonesOutput{
					AvailabilityZones: []*awsec2.AvailabilityZone{{ZoneName: nil}},
				}

				_, err := client.RetrieveAvailabilityZones("us-east-1")
				Expect(err).To(MatchError("aws returned availability zone with nil zone name"))
			})

			It("returns an error when describe availability zones fails", func() {
				ec2Client.DescribeAvailabilityZonesCall.Returns.Error = errors.New("describe availability zones failed")
				_, err := client.RetrieveAvailabilityZones("us-east-1")
				Expect(err).To(MatchError("describe availability zones failed"))
			})
		})
	})

	Describe("DeleteKeyPair", func() {
		var (
			client    ec2.Client
			ec2Client *fakes.AWSEC2Client
			logger    *fakes.Logger
		)

		BeforeEach(func() {
			ec2Client = &fakes.AWSEC2Client{}
			logger = &fakes.Logger{}
			client = ec2.NewClientWithInjectedEC2Client(ec2Client, logger)
		})

		It("deletes the ec2 keypair", func() {
			err := client.DeleteKeyPair("some-key-pair-name")
			Expect(err).NotTo(HaveOccurred())

			Expect(ec2Client.DeleteKeyPairCall.Receives.Input).To(Equal(&awsec2.DeleteKeyPairInput{
				KeyName: awslib.String("some-key-pair-name"),
			}))

			Expect(logger.StepCall.Receives.Message).To(Equal("deleting keypair"))
		})

		Context("when the keypair cannot be deleted", func() {
			It("returns an error", func() {
				ec2Client.DeleteKeyPairCall.Returns.Error = errors.New("failed to delete keypair")

				err := client.DeleteKeyPair("some-key-pair-name")
				Expect(err).To(MatchError("failed to delete keypair"))
			})
		})
	})

	Describe("ValidateSafeToDelete", func() {
		var (
			client    ec2.Client
			ec2Client *fakes.AWSEC2Client
		)

		BeforeEach(func() {
			ec2Client = &fakes.AWSEC2Client{}
			client = ec2.NewClientWithInjectedEC2Client(ec2Client, &fakes.Logger{})
		})

		It("returns nil when the only EC2 instances are bosh and nat", func() {
			ec2Client.DescribeInstancesCall.Returns.Output = &awsec2.DescribeInstancesOutput{
				Reservations: []*awsec2.Reservation{
					reservationContainingInstance("NAT"),
					reservationContainingInstance("bosh/0"),
				},
			}

			err := client.ValidateSafeToDelete("some-vpc-id", "")
			Expect(err).NotTo(HaveOccurred())

			Expect(ec2Client.DescribeInstancesCall.Receives.Input).To(Equal(&awsec2.DescribeInstancesInput{
				Filters: []*awsec2.Filter{{
					Name:   awslib.String("vpc-id"),
					Values: []*string{awslib.String("some-vpc-id")},
				}},
			}))
		})

		Context("when passed an environment ID", func() {
			It("returns nil when the only EC2 instances are bosh, jumpbox and envID-nat", func() {
				ec2Client.DescribeInstancesCall.Returns.Output = &awsec2.DescribeInstancesOutput{
					Reservations: []*awsec2.Reservation{
						reservationContainingInstance("example-env-id-nat"),
						reservationContainingInstance("bosh/0"),
						reservationContainingInstance("jumpbox/0"),
					},
				}

				err := client.ValidateSafeToDelete("some-vpc-id", "example-env-id")
				Expect(err).NotTo(HaveOccurred())

				Expect(ec2Client.DescribeInstancesCall.Receives.Input).To(Equal(&awsec2.DescribeInstancesInput{
					Filters: []*awsec2.Filter{{
						Name:   awslib.String("vpc-id"),
						Values: []*string{awslib.String("some-vpc-id")},
					}},
				}))
			})
		})

		It("returns nil when there are no instances at all", func() {
			ec2Client.DescribeInstancesCall.Returns.Output = &awsec2.DescribeInstancesOutput{
				Reservations: []*awsec2.Reservation{},
			}

			err := client.ValidateSafeToDelete("some-vpc-id", "")
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

			err := client.ValidateSafeToDelete("some-vpc-id", "")
			Expect(err).To(MatchError("vpc some-vpc-id is not safe to delete; vms still exist: [first-bosh-deployed-vm, second-bosh-deployed-vm]"))
		})

		It("returns an error even when there are two VMs in the VPC, but they are not NAT and BOSH", func() {
			ec2Client.DescribeInstancesCall.Returns.Output = &awsec2.DescribeInstancesOutput{
				Reservations: []*awsec2.Reservation{
					reservationContainingInstance("not-bosh"),
					reservationContainingInstance("not-nat"),
				},
			}

			err := client.ValidateSafeToDelete("some-vpc-id", "")
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

			err := client.ValidateSafeToDelete("some-vpc-id", "")
			Expect(err).To(MatchError("vpc some-vpc-id is not safe to delete; vms still exist: [NAT, bosh/0, bosh/0]"))
		})

		It("returns an error even if the vpc contains untagged vms", func() {
			ec2Client.DescribeInstancesCall.Returns.Output = &awsec2.DescribeInstancesOutput{
				Reservations: []*awsec2.Reservation{
					&awsec2.Reservation{
						Instances: []*awsec2.Instance{{
							Tags: []*awsec2.Tag{{
								Key:   awslib.String("Name"),
								Value: awslib.String(""),
							}},
						}, {}, {}},
					},
				},
			}

			err := client.ValidateSafeToDelete("some-vpc-id", "")
			Expect(err).To(MatchError("vpc some-vpc-id is not safe to delete; vms still exist: [unnamed, unnamed, unnamed]"))
		})

		Describe("failure cases", func() {
			It("returns an error when the describe instances call fails", func() {
				ec2Client.DescribeInstancesCall.Returns.Error = errors.New("failed to describe instances")
				err := client.ValidateSafeToDelete("some-vpc-id", "")
				Expect(err).To(MatchError("failed to describe instances"))
			})
		})
	})
})

func reservationContainingInstance(tag string) *awsec2.Reservation {
	return &awsec2.Reservation{
		Instances: []*awsec2.Instance{{
			Tags: []*awsec2.Tag{{
				Key:   awslib.String("Name"),
				Value: awslib.String(tag),
			}},
		}},
	}
}
