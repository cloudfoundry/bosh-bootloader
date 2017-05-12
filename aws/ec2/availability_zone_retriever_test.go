package ec2_test

import (
	"errors"

	goaws "github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cloudfoundry/bosh-bootloader/aws/ec2"
	"github.com/cloudfoundry/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AvailabilityZoneRetriever", func() {
	var (
		availabilityZoneRetriever ec2.AvailabilityZoneRetriever
		ec2Client                 *fakes.EC2Client
		awsClientProvider         *fakes.AWSClientProvider
	)

	BeforeEach(func() {
		ec2Client = &fakes.EC2Client{}
		awsClientProvider = &fakes.AWSClientProvider{}
		awsClientProvider.GetEC2ClientCall.Returns.EC2Client = ec2Client
		availabilityZoneRetriever = ec2.NewAvailabilityZoneRetriever(awsClientProvider)
	})

	It("fetches availability zones for a given region", func() {
		ec2Client.DescribeAvailabilityZonesCall.Returns.Output = &awsec2.DescribeAvailabilityZonesOutput{
			AvailabilityZones: []*awsec2.AvailabilityZone{
				{ZoneName: goaws.String("us-east-1a")},
				{ZoneName: goaws.String("us-east-1b")},
				{ZoneName: goaws.String("us-east-1c")},
				{ZoneName: goaws.String("us-east-1e")},
			},
		}

		azs, err := availabilityZoneRetriever.Retrieve("us-east-1")

		Expect(err).NotTo(HaveOccurred())
		Expect(azs).To(ConsistOf("us-east-1a", "us-east-1b", "us-east-1c", "us-east-1e"))
		Expect(ec2Client.DescribeAvailabilityZonesCall.Receives.Input).To(Equal(&awsec2.DescribeAvailabilityZonesInput{
			Filters: []*awsec2.Filter{{
				Name:   goaws.String("region-name"),
				Values: []*string{goaws.String("us-east-1")},
			}},
		}))
	})

	Describe("failure cases", func() {
		It("returns an error when AWS returns a nil availability zone", func() {
			ec2Client.DescribeAvailabilityZonesCall.Returns.Output = &awsec2.DescribeAvailabilityZonesOutput{
				AvailabilityZones: []*awsec2.AvailabilityZone{nil},
			}

			_, err := availabilityZoneRetriever.Retrieve("us-east-1")
			Expect(err).To(MatchError("aws returned nil availability zone"))
		})

		It("returns an error when an availability zone with a nil ZoneName", func() {
			ec2Client.DescribeAvailabilityZonesCall.Returns.Output = &awsec2.DescribeAvailabilityZonesOutput{
				AvailabilityZones: []*awsec2.AvailabilityZone{{ZoneName: nil}},
			}

			_, err := availabilityZoneRetriever.Retrieve("us-east-1")
			Expect(err).To(MatchError("aws returned availability zone with nil zone name"))
		})

		It("returns an error when describe availability zones fails", func() {
			ec2Client.DescribeAvailabilityZonesCall.Returns.Error = errors.New("describe availability zones failed")
			_, err := availabilityZoneRetriever.Retrieve("us-east-1")
			Expect(err).To(MatchError("describe availability zones failed"))
		})
	})
})
