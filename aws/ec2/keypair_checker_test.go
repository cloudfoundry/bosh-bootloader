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

var _ = Describe("KeyPairChecker", func() {
	var (
		ec2Client         *fakes.EC2Client
		checker           ec2.KeyPairChecker
		awsClientProvider *fakes.AWSClientProvider
	)

	BeforeEach(func() {
		ec2Client = &fakes.EC2Client{}
		awsClientProvider = &fakes.AWSClientProvider{}
		awsClientProvider.GetEC2ClientCall.Returns.EC2Client = ec2Client
		checker = ec2.NewKeyPairChecker(awsClientProvider)
	})

	Describe("HasKeyPair", func() {
		Context("when the keypair exists on AWS", func() {
			BeforeEach(func() {
				ec2Client.DescribeKeyPairsCall.Returns.Output = &awsec2.DescribeKeyPairsOutput{
					KeyPairs: []*awsec2.KeyPairInfo{
						{
							KeyFingerprint: goaws.String("some-finger-print"),
							KeyName:        goaws.String("some-key-name"),
						},
					},
				}
			})

			It("returns true", func() {
				present, err := checker.HasKeyPair("some-key-name")
				Expect(err).NotTo(HaveOccurred())
				Expect(present).To(BeTrue())
				Expect(awsClientProvider.GetEC2ClientCall.CallCount).To(Equal(1))

				Expect(ec2Client.DescribeKeyPairsCall.Receives.Input).To(Equal(&awsec2.DescribeKeyPairsInput{
					KeyNames: []*string{
						goaws.String("some-key-name"),
					},
				}))
			})
		})

		Context("when the keypair does not exist on AWS", func() {
			It("returns false when the keypair name can not be found", func() {
				ec2Client.DescribeKeyPairsCall.Returns.Error = errors.New("InvalidKeyPair.NotFound")

				present, err := checker.HasKeyPair("some-key-name")
				Expect(err).NotTo(HaveOccurred())
				Expect(present).To(BeFalse())
			})

			It("returns false when the keypair name is empty", func() {
				ec2Client.DescribeKeyPairsCall.Returns.Error = errors.New("InvalidParameterValue: Invalid value '' for keyPairNames. It should not be blank")

				present, err := checker.HasKeyPair("")
				Expect(err).NotTo(HaveOccurred())
				Expect(present).To(BeFalse())
			})
		})

		Context("failure cases", func() {
			It("returns an error when AWS communication fails", func() {
				ec2Client.DescribeKeyPairsCall.Returns.Error = errors.New("something bad happened")

				_, err := checker.HasKeyPair("some-key-name")
				Expect(err).To(MatchError("something bad happened"))
			})
		})
	})
})
