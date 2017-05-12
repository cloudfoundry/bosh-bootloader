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

var _ = Describe("KeyPairDeleter", func() {
	var (
		deleter           ec2.KeyPairDeleter
		client            *fakes.EC2Client
		awsClientProvider *fakes.AWSClientProvider
		logger            *fakes.Logger
	)

	BeforeEach(func() {
		awsClientProvider = &fakes.AWSClientProvider{}
		client = &fakes.EC2Client{}
		logger = &fakes.Logger{}
		awsClientProvider.GetEC2ClientCall.Returns.EC2Client = client
		deleter = ec2.NewKeyPairDeleter(awsClientProvider, logger)
	})

	It("deletes the ec2 keypair", func() {
		err := deleter.Delete("some-key-pair-name")
		Expect(err).NotTo(HaveOccurred())

		Expect(client.DeleteKeyPairCall.Receives.Input).To(Equal(&awsec2.DeleteKeyPairInput{
			KeyName: aws.String("some-key-pair-name"),
		}))

		Expect(logger.StepCall.Receives.Message).To(Equal("deleting keypair"))
	})

	Context("failure cases", func() {
		Context("when the keypair cannot be deleted", func() {
			It("returns an error", func() {
				client.DeleteKeyPairCall.Returns.Error = errors.New("failed to delete keypair")

				err := deleter.Delete("some-key-pair-name")
				Expect(err).To(MatchError("failed to delete keypair"))
			})
		})
	})
})
