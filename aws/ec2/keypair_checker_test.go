package ec2_test

import (
	"errors"

	goaws "github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("KeyPairChecker", func() {
	var (
		ec2Client *fakes.EC2Client
		checker   ec2.KeyPairChecker
	)

	BeforeEach(func() {
		ec2Client = &fakes.EC2Client{}
		checker = ec2.NewKeyPairChecker()
	})

	Describe("HasKeyPair", func() {
		Context("when the keypair exists on AWS", func() {
			BeforeEach(func() {
				ec2Client.DescribeKeyPairsCall.Returns.DescribeKeyPairsOutput = &awsec2.DescribeKeyPairsOutput{
					KeyPairs: []*awsec2.KeyPairInfo{
						{
							KeyFingerprint: goaws.String("some-finger-print"),
							KeyName:        goaws.String("some-key-name"),
						},
					},
				}
			})

			It("returns true", func() {
				present, err := checker.HasKeyPair(ec2Client, "some-key-name")
				Expect(err).NotTo(HaveOccurred())
				Expect(present).To(BeTrue())

				Expect(ec2Client.DescribeKeyPairsCall.Receives.Input).To(Equal(&awsec2.DescribeKeyPairsInput{
					KeyNames: []*string{
						goaws.String("some-key-name"),
					},
				}))
			})
		})

		Context("when the keypair does not exist on AWS", func() {
			It("returns false", func() {
				ec2Client.DescribeKeyPairsCall.Returns.Error = errors.New("InvalidKeyPair.NotFound")

				present, err := checker.HasKeyPair(ec2Client, "some-key-name")
				Expect(err).NotTo(HaveOccurred())
				Expect(present).To(BeFalse())
			})
		})

		Context("failure cases", func() {
			It("returns an error when AWS communication fails", func() {
				ec2Client.DescribeKeyPairsCall.Returns.Error = errors.New("something bad happened")

				_, err := checker.HasKeyPair(ec2Client, "some-key-name")
				Expect(err).To(MatchError("something bad happened"))
			})
		})
	})
})
