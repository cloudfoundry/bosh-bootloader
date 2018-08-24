package ec2_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/genevieve/leftovers/aws/ec2"
	"github.com/genevieve/leftovers/aws/ec2/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("KeyPair", func() {
	var (
		keyPair ec2.KeyPair
		client  *fakes.KeyPairsClient
		name    *string
	)

	BeforeEach(func() {
		client = &fakes.KeyPairsClient{}
		name = aws.String("the-name")

		keyPair = ec2.NewKeyPair(client, name)
	})

	Describe("Delete", func() {
		It("deletes the key pair", func() {
			err := keyPair.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteKeyPairCall.CallCount).To(Equal(1))
			Expect(client.DeleteKeyPairCall.Receives.Input.KeyName).To(Equal(name))
		})

		Context("when the client fails", func() {
			BeforeEach(func() {
				client.DeleteKeyPairCall.Returns.Error = awserr.New("InvalidKeyPair.NotFound", "", nil)
			})

			It("returns nil", func() {
				err := keyPair.Delete()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the key is not found", func() {
			BeforeEach(func() {
				client.DeleteKeyPairCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := keyPair.Delete()
				Expect(err).To(MatchError("Delete: banana"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the identifier", func() {
			Expect(keyPair.Name()).To(Equal("the-name"))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(keyPair.Type()).To(Equal("EC2 Key Pair"))
		})
	})
})
