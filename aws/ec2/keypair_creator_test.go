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
	)

	var fakeUUIDGenerator = func() (string, error) {
		return "guid", nil
	}

	BeforeEach(func() {
		ec2Client = &fakes.EC2Client{}
		keyPairCreator = ec2.NewKeyPairCreator(fakeUUIDGenerator)
	})

	Describe("Create", func() {
		It("creates a new keypair on ec2", func() {
			ec2Client.CreateKeyPairCall.Returns.Output = &awsec2.CreateKeyPairOutput{
				KeyFingerprint: goaws.String("some-fingerprint"),
				KeyMaterial:    goaws.String("some-private-key"),
				KeyName:        goaws.String("keypair-guid"),
			}

			keyPair, err := keyPairCreator.Create(ec2Client)
			Expect(err).NotTo(HaveOccurred())

			Expect(keyPair.Name).To(Equal("keypair-guid"))
			Expect(keyPair.PrivateKey).To(Equal([]byte("some-private-key")))

			Expect(ec2Client.CreateKeyPairCall.Receives.Input).To(Equal(&awsec2.CreateKeyPairInput{
				KeyName: goaws.String("keypair-guid"),
			}))
		})

		Context("failure cases", func() {
			Context("when a guid cannot be generated", func() {
				It("returns an error", func() {
					erroringUUIDGenerator := func() (string, error) {
						return "", errors.New("failed to generate guid")
					}
					keyPairCreator = ec2.NewKeyPairCreator(erroringUUIDGenerator)

					_, err := keyPairCreator.Create(ec2Client)
					Expect(err).To(MatchError("failed to generate guid"))
				})
			})

			Context("when the create keypair request fails", func() {
				It("returns an error", func() {
					ec2Client.CreateKeyPairCall.Returns.Error = errors.New("failed to create keypair")

					_, err := keyPairCreator.Create(ec2Client)
					Expect(err).To(MatchError("failed to create keypair"))
				})
			})
		})
	})
})
