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

var _ = Describe("KeyPairCreator", func() {
	var (
		keyPairCreator ec2.KeyPairCreator
		ec2Client      *fakes.EC2Client
		clientProvider *fakes.ClientProvider
	)

	BeforeEach(func() {
		clientProvider = &fakes.ClientProvider{}
		ec2Client = &fakes.EC2Client{}
		clientProvider.GetEC2ClientCall.Returns.EC2Client = ec2Client
		keyPairCreator = ec2.NewKeyPairCreator(clientProvider)
	})

	Describe("Create", func() {
		It("creates a new keypair on ec2", func() {
			ec2Client.CreateKeyPairCall.Returns.Output = &awsec2.CreateKeyPairOutput{
				KeyFingerprint: goaws.String("some-fingerprint"),
				KeyMaterial:    goaws.String("some-private-key"),
				KeyName:        goaws.String("keypair-guid"),
			}

			keyPair, err := keyPairCreator.Create("keypair-some-env-id")
			Expect(err).NotTo(HaveOccurred())

			Expect(keyPair.Name).To(Equal("keypair-some-env-id"))
			Expect(keyPair.PrivateKey).To(Equal("some-private-key"))

			Expect(ec2Client.CreateKeyPairCall.Receives.Input).To(Equal(&awsec2.CreateKeyPairInput{
				KeyName: goaws.String("keypair-some-env-id"),
			}))
		})

		Context("failure cases", func() {
			Context("when the create keypair request fails", func() {
				It("returns an error", func() {
					ec2Client.CreateKeyPairCall.Returns.Error = errors.New("failed to create keypair")

					_, err := keyPairCreator.Create("")
					Expect(err).To(MatchError("failed to create keypair"))
				})
			})
		})
	})
})
