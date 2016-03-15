package ec2_test

import (
	"errors"
	"fmt"

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
		uuidGenerator  *fakes.UUIDGenerator
	)

	BeforeEach(func() {
		ec2Client = &fakes.EC2Client{}
		uuidGenerator = &fakes.UUIDGenerator{}
		keyPairCreator = ec2.NewKeyPairCreator(uuidGenerator)
	})

	Describe("Create", func() {
		It("creates a new keypair on ec2", func() {
			ec2Client.CreateKeyPairCall.Returns.Output = &awsec2.CreateKeyPairOutput{
				KeyFingerprint: goaws.String("some-fingerprint"),
				KeyMaterial:    goaws.String("some-private-key"),
				KeyName:        goaws.String("keypair-guid"),
			}
			uuidGenerator.GenerateCall.Returns = []fakes.GenerateReturn{{String: "guid"}}

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
				BeforeEach(func() {
					uuidGenerator.GenerateCall.Returns = []fakes.GenerateReturn{{Error: fmt.Errorf("failed to generate guid")}}
				})

				It("returns an error", func() {
					keyPairCreator = ec2.NewKeyPairCreator(uuidGenerator)

					_, err := keyPairCreator.Create(ec2Client)
					Expect(err).To(MatchError("failed to generate guid"))
				})
			})

			Context("when the create keypair request fails", func() {
				BeforeEach(func() {
					uuidGenerator.GenerateCall.Returns = []fakes.GenerateReturn{fakes.GenerateReturn{}}
				})
				It("returns an error", func() {
					ec2Client.CreateKeyPairCall.Returns.Error = errors.New("failed to create keypair")

					_, err := keyPairCreator.Create(ec2Client)
					Expect(err).To(MatchError("failed to create keypair"))
				})
			})
		})
	})
})
